package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yunhu-channel/yunhu-channel/internal/config"
	"github.com/yunhu-channel/yunhu-channel/internal/jsonrpc"
	"github.com/yunhu-channel/yunhu-channel/internal/webhook"
	"github.com/yunhu-channel/yunhu-channel/internal/yunhu"
)

const shutdownTimeout = 5 * time.Second

type InboundMessage struct {
	Message InboundMessagePayload `json:"message"`
}

type InboundMessagePayload struct {
	SenderID  string                 `json:"sender_id"`
	ChatID    string                 `json:"chat_id"`
	Text      string                 `json:"text"`
	Media     []string               `json:"media"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type Handler struct {
	transport   *jsonrpc.StdioTransport
	config      *config.Config
	runtime     *config.Runtime
	yunhuClient *yunhu.Client
	webhookSrv  *webhook.Server

	started   bool
	chatTypes map[string]string
	streamBuf map[string]string
	mu        sync.Mutex
}

func NewHandler(transport *jsonrpc.StdioTransport) *Handler {
	return &Handler{
		transport: transport,
		chatTypes: make(map[string]string),
		streamBuf: make(map[string]string),
	}
}

func (h *Handler) Handle(ctx context.Context, req *jsonrpc.Request) {
	if req == nil || !req.HasID() {
		return
	}

	var resp *jsonrpc.Response
	var err error

	switch req.Method {
	case "get_manifest":
		resp, err = h.handleGetManifest(req)
	case "start":
		resp, err = h.handleStart(req)
	case "stop":
		resp, err = h.handleStop(req)
	case "send":
		resp, err = h.handleSend(req)
	case "send_rich":
		resp, err = h.handleSendRich(req)
	case "edit_message":
		resp, err = h.handleEditMessage(req)
	case "delete_message":
		resp, err = h.handleDeleteMessage(req)
	case "health":
		resp, err = h.handleHealth(req)
	default:
		resp = jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeMethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
	}

	if err != nil {
		slog.Error("handler error", "method", req.Method, "error", err)
		resp = jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInternalError, err.Error())
	}

	if resp != nil {
		if writeErr := h.transport.WriteResponse(resp); writeErr != nil {
			slog.Error("failed to write response", "error", writeErr)
		}
	}
}

func (h *Handler) handleGetManifest(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	return jsonrpc.NewResponse(req.ID, BuildManifest())
}

func (h *Handler) handleStart(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return jsonrpc.NewResponse(req.ID, StartResult{Started: true})
	}

	sp, err := config.ParseStartParams(req.Params)
	if err != nil {
		return nil, fmt.Errorf("parse start params: %w", err)
	}

	cfg, err := sp.ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	h.config = cfg
	h.runtime = &sp.Runtime
	h.yunhuClient = yunhu.NewClient(cfg.Token)

	if h.runtime.StateDir != "" {
		if err := os.MkdirAll(h.runtime.StateDir, 0700); err != nil {
			slog.Warn("failed to create state directory", "path", h.runtime.StateDir, "error", err)
		} else {
			pidFile := filepath.Join(h.runtime.StateDir, "yunhu-channel.pid")
			pidData := fmt.Sprintf("%d\n", os.Getpid())
			if err := os.WriteFile(pidFile, []byte(pidData), 0600); err != nil {
				slog.Warn("failed to write pid file", "path", pidFile, "error", err)
			}
			slog.Info("state directory initialized", "path", h.runtime.StateDir)
		}
	}

	cfg.LogInfo()

	inboundCh := make(chan []byte, 128)
	h.webhookSrv = webhook.NewServer(cfg, h.runtime, func(notification []byte) {
		select {
		case inboundCh <- notification:
		default:
			slog.Error("inbound notification channel full, dropping message")
		}
	})

	if err := h.webhookSrv.Start(); err != nil {
		return nil, fmt.Errorf("start webhook server: %w", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in inbound notification goroutine", "recover", r)
			}
		}()
		for notification := range inboundCh {
			h.trackChatType(notification)
			if err := h.transport.WriteRaw(notification); err != nil {
				slog.Error("failed to write inbound notification", "error", err)
			}
		}
	}()

	h.started = true
	return jsonrpc.NewResponse(req.ID, StartResult{Started: true})
}

func (h *Handler) handleStop(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.webhookSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := h.webhookSrv.Shutdown(ctx); err != nil {
			slog.Error("webhook shutdown error", "error", err)
		}
		h.webhookSrv = nil
	}

	h.started = false
	return jsonrpc.NewResponse(req.ID, StopResult{Accepted: true})
}

func (h *Handler) handleSend(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()

	if !h.started || h.yunhuClient == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("channel not started")
	}

	var params struct {
		Message struct {
			Target string          `json:"target"`
			Text   string          `json:"text"`
			Stage  string          `json:"stage"`
			Media  []string        `json:"media"`
		} `json:"message"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("parse send params: %w", err)
	}

	if params.Message.Target == "" {
		h.mu.Unlock()
		return nil, fmt.Errorf("target is required")
	}

	// Handle streaming chunk (just buffer, no HTTP call)
	if params.Message.Stage == "chunk" {
		h.streamBuf[params.Message.Target] += params.Message.Text
		h.mu.Unlock()
		return jsonrpc.NewResponse(req.ID, SendResult{Accepted: true})
	}

	// Merge buffered streaming content
	if params.Message.Stage == "final" && h.streamBuf[params.Message.Target] != "" {
		h.streamBuf[params.Message.Target] += params.Message.Text
		params.Message.Text = h.streamBuf[params.Message.Target]
		delete(h.streamBuf, params.Message.Target)
	}

	// Read recvType directly while holding the lock (avoid re-entrant lock)
	recvType := yunhu.RecvTypeUser
	if t, ok := h.chatTypes[params.Message.Target]; ok {
		recvType = t
	}
	client := h.yunhuClient
	target := params.Message.Target
	text := params.Message.Text
	media := params.Message.Media
	h.mu.Unlock()

	// HTTP call without holding the lock, so health checks are not blocked
	if len(media) == 0 {
		sendReq := &yunhu.SendMessageRequest{
			RecvID:      target,
			RecvType:    recvType,
			ContentType: yunhu.ContentTypeMarkdown,
			Content:     yunhu.SendContent{Text: text},
		}
		_, err := client.SendMessage(sendReq)
		if err != nil {
			slog.Error("send failed", "target", target, "error", err)
			return jsonrpc.NewResponse(req.ID, SendResult{Accepted: false})
		}
		return jsonrpc.NewResponse(req.ID, SendResult{Accepted: true})
	}

	contentType, sendContent := h.buildSendContent(text, media)
	msgReq := &yunhu.SendMessageRequest{
		RecvID:      target,
		RecvType:    recvType,
		ContentType: contentType,
		Content:     sendContent,
	}

	_, err := client.SendMessage(msgReq)
	if err != nil {
		slog.Error("send message failed", "target", target, "error", err)
		return jsonrpc.NewResponse(req.ID, SendResult{Accepted: false})
	}

	return jsonrpc.NewResponse(req.ID, SendResult{Accepted: true})
}

func (h *Handler) handleSendRich(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()

	if !h.started || h.yunhuClient == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("channel not started")
	}

	var params struct {
		Message struct {
			Target      string `json:"target"`
			Text        string `json:"text"`
			Attachments []interface{} `json:"attachments"`
			Choices     []struct {
				ID         string `json:"id"`
				Label      string `json:"label"`
				SubmitText string `json:"submit_text"`
			} `json:"choices"`
		} `json:"message"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("parse send_rich params: %w", err)
	}

	if params.Message.Target == "" {
		h.mu.Unlock()
		return nil, fmt.Errorf("target is required")
	}

	var buttons []yunhu.Button
	for _, choice := range params.Message.Choices {
		buttons = append(buttons, yunhu.Button{
			Text:       choice.Label,
			ActionType: yunhu.ButtonActionReport,
			Value:      choice.SubmitText,
		})
	}

	recvType := yunhu.RecvTypeUser
	if t, ok := h.chatTypes[params.Message.Target]; ok {
		recvType = t
	}
	client := h.yunhuClient
	target := params.Message.Target
	text := params.Message.Text
	h.mu.Unlock()

	msgReq := &yunhu.SendMessageRequest{
		RecvID:      target,
		RecvType:    recvType,
		ContentType: yunhu.ContentTypeMarkdown,
		Content: yunhu.SendContent{
			Text:    text,
			Buttons: buttons,
		},
	}

	sendResp, err := client.SendMessage(msgReq)
	if err != nil {
		slog.Error("send rich message failed", "target", params.Message.Target, "error", err)
		result := SendResult{Accepted: false}
		return jsonrpc.NewResponse(req.ID, result)
	}

	result := SendResult{Accepted: true}
	if sendResp.Data != nil && sendResp.Data.MessageInfo != nil {
		result.MessageID = sendResp.Data.MessageInfo.MsgID
	}
	return jsonrpc.NewResponse(req.ID, result)
}

func (h *Handler) handleEditMessage(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()

	if !h.started || h.yunhuClient == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("channel not started")
	}

	var params struct {
		Message struct {
			Target    string `json:"target"`
			MessageID string `json:"message_id"`
			Text      string `json:"text"`
		} `json:"message"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("parse edit_message params: %w", err)
	}

	if params.Message.Target == "" || params.Message.MessageID == "" {
		h.mu.Unlock()
		return nil, fmt.Errorf("target and message_id are required")
	}

	// Read recvType while holding lock, then unlock before HTTP call
	recvType := yunhu.RecvTypeUser
	if t, ok := h.chatTypes[params.Message.Target]; ok {
		recvType = t
	}
	client := h.yunhuClient
	target := params.Message.Target
	msgID := params.Message.MessageID
	text := params.Message.Text
	h.mu.Unlock()

	editReq := &yunhu.EditMessageRequest{
		MsgID:       msgID,
		RecvID:      target,
		RecvType:    recvType,
		ContentType: yunhu.ContentTypeMarkdown,
		Content: yunhu.SendContent{
			Text: text,
		},
	}

	_, err := client.EditMessage(editReq)
	if err != nil {
		slog.Error("edit message failed", "message_id", params.Message.MessageID, "error", err)
		return jsonrpc.NewResponse(req.ID, EditResult{Accepted: false})
	}

	return jsonrpc.NewResponse(req.ID, EditResult{Accepted: true})
}

func (h *Handler) handleDeleteMessage(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()

	if !h.started || h.yunhuClient == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("channel not started")
	}

	var params struct {
		Message struct {
			Target    string `json:"target"`
			MessageID string `json:"message_id"`
		} `json:"message"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("parse delete_message params: %w", err)
	}

	if params.Message.Target == "" || params.Message.MessageID == "" {
		h.mu.Unlock()
		return nil, fmt.Errorf("target and message_id are required")
	}

	chatType := yunhu.ChatTypeBot
	if t, ok := h.chatTypes[params.Message.Target]; ok && t == yunhu.RecvTypeGroup {
		chatType = yunhu.ChatTypeGroup
	}

	client := h.yunhuClient
	target := params.Message.Target
	msgID := params.Message.MessageID
	h.mu.Unlock()

	_, err := client.RecallMessage(&yunhu.RecallMessageRequest{
		MsgID:    msgID,
		ChatType: chatType,
		ChatID:   target,
	})
	if err != nil {
		slog.Error("delete message failed", "message_id", params.Message.MessageID, "error", err)
		return jsonrpc.NewResponse(req.ID, DeleteResult{Accepted: false})
	}

	return jsonrpc.NewResponse(req.ID, DeleteResult{Accepted: true})
}

func (h *Handler) handleHealth(req *jsonrpc.Request) (*jsonrpc.Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	healthy := true

	if h.config == nil || h.config.Token == "" {
		healthy = false
	}

	slog.Info("health check", "healthy", healthy, "webhook_running", h.webhookSrv != nil && h.webhookSrv.IsRunning(), "token_configured", h.config != nil && h.config.Token != "")

	return jsonrpc.NewResponse(req.ID, HealthResult{Healthy: healthy})
}

func (h *Handler) Shutdown(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.webhookSrv != nil {
		if err := h.webhookSrv.Shutdown(ctx); err != nil {
			return err
		}
		h.webhookSrv = nil
	}
	h.started = false
	return nil
}

func (h *Handler) trackChatType(notification []byte) {
	var notif struct {
		Params struct {
			Message struct {
				ChatID   string                 `json:"chat_id"`
				Metadata map[string]interface{} `json:"metadata"`
			} `json:"message"`
		} `json:"params"`
	}
	if err := json.Unmarshal(notification, &notif); err != nil {
		return
	}
	msg := notif.Params.Message
	if msg.ChatID == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if msg.Metadata != nil {
		if pk, ok := msg.Metadata["peer_kind"].(string); ok {
			if pk == "group" {
				h.chatTypes[msg.ChatID] = yunhu.RecvTypeGroup
				return
			}
		}
	}
	h.chatTypes[msg.ChatID] = yunhu.RecvTypeUser
}

func (h *Handler) inferRecvType(target string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if t, ok := h.chatTypes[target]; ok {
		return t
	}
	return yunhu.RecvTypeUser
}

// buildSendContent 根据 text 和 media 决定消息的 contentType 和内容。
//
// media 中的每一项可以是：
//   - 已上传的文件 key（不含 "://"）— 用于图片/视频/文件类型的消息
//   - 公开 URL（含 "://"）— 以 markdown 图片形式嵌入到文本中
//
// 图片/视频/文件需要先调用 Upload 接口拿到 key，
// 不能直接传 URL 给 imageKey/videoKey/fileKey。
func (h *Handler) buildSendContent(text string, media []string) (string, yunhu.SendContent) {
	if len(media) == 0 {
		return yunhu.ContentTypeMarkdown, yunhu.SendContent{Text: text}
	}

	firstMedia := media[0]
	if text == "" && len(media) == 1 {
		if !strings.Contains(firstMedia, "://") {
			return yunhu.ContentTypeImage, yunhu.SendContent{ImageKey: firstMedia}
		}
	}

	content := text
	for _, m := range media {
		content += "\n![](" + m + ")"
	}
	return yunhu.ContentTypeMarkdown, yunhu.SendContent{Text: content}
}
