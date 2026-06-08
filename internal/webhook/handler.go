package webhook

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/yunhu-channel/yunhu-channel/internal/config"
)

type EventMsgVo struct {
	Version string `json:"version"`
	Header  struct {
		EventID   string `json:"eventId"`
		EventTime int64  `json:"eventTime"`
		EventType string `json:"eventType"`
	} `json:"header"`
	Event struct {
		Time      int64  `json:"time"`
		ChatID    string `json:"chatId"`
		ChatType  string `json:"chatType"`
		GroupID   string `json:"groupId"`
		GroupName string `json:"groupName"`
		UserID    string `json:"userId"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`

		Sender struct {
			SenderID        string `json:"senderId"`
			SenderType      string `json:"senderType"`
			SenderUserLevel string `json:"senderUserLevel"`
			SenderNickname  string `json:"senderNickname"`
		} `json:"sender"`

		Chat struct {
			ChatID   string `json:"chatId"`
			ChatType string `json:"chatType"`
		} `json:"chat"`

		Message struct {
			MsgID       string `json:"msgId"`
			ParentID    string `json:"parentId"`
			SendTime    int64  `json:"sendTime"`
			ChatID      string `json:"chatId"`
			ChatType    string `json:"chatType"`
			ContentType string `json:"contentType"`
			CommandID   int    `json:"commandId"`
			CommandName string `json:"commandName"`
			Content     struct {
				Text      string `json:"text"`
				ImageURL  string `json:"imageUrl"`
				ImageName string `json:"imageName"`
				FileName  string `json:"fileName"`
				FileURL   string `json:"fileUrl"`
				FileSize  int64  `json:"fileSize"`
			} `json:"content"`
		} `json:"message"`

		RecvID   string `json:"recvId"`
		RecvType string `json:"recvType"`
		Value    string `json:"value"`
	} `json:"event"`
}

type InboundNotification struct {
	Message InboundMessagePayload `json:"message"`
}

type InboundMessagePayload struct {
	SenderID string                 `json:"sender_id"`
	ChatID   string                 `json:"chat_id"`
	Text     string                 `json:"text"`
	Media    []string               `json:"media"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MarshalJSON ensures Media serializes as [] not null when empty.
func (p InboundMessagePayload) MarshalJSON() ([]byte, error) {
	type Alias InboundMessagePayload
	a := Alias(p)
	if a.Media == nil {
		a.Media = []string{}
	}
	return json.Marshal(a)
}

type WebhookHandler struct {
	config   *config.Config
	runtime  *config.Runtime
	notifyFn func([]byte)
}

func NewWebhookHandler(cfg *config.Config, rt *config.Runtime, notifyFn func([]byte)) *WebhookHandler {
	return &WebhookHandler{
		config:   cfg,
		runtime:  rt,
		notifyFn: notifyFn,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		slog.Error("failed to read webhook body", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var event EventMsgVo
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("failed to parse webhook event", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	inbound := h.buildInboundMessage(&event)
	if inbound == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":0}`))

	notification, err := json.Marshal(&inboundNotification{
		JSONRPC: "2.0",
		Method:  "inbound_message",
		Params:  inbound,
	})
	if err != nil {
		slog.Error("failed to marshal inbound notification", "error", err)
		return
	}

	h.notifyFn(notification)
}

func (h *WebhookHandler) buildInboundMessage(event *EventMsgVo) *InboundNotification {
	eventType := event.Header.EventType

	switch eventType {
	case "message.receive.normal":
		return h.buildNormalMessage(event)
	case "message.receive.instruction":
		return h.buildInstructionMessage(event)
	case "button.report.inline":
		return h.buildButtonReport(event)
	case "group.join":
		return h.buildGroupJoin(event)
	case "group.leave":
		return h.buildGroupLeave(event)
	case "bot.followed":
		return h.buildBotFollowed(event)
	case "bot.unfollowed":
		return h.buildBotUnfollowed(event)
	default:
		slog.Debug("unhandled event type", "eventType", eventType)
		return nil
	}
}

func (h *WebhookHandler) buildNormalMessage(event *EventMsgVo) *InboundNotification {
	text := event.Event.Message.Content.Text
	if text == "" && event.Event.Message.Content.ImageURL != "" {
		text = " [图片]"
	}
	if text == "" && event.Event.Message.Content.FileURL != "" {
		text = " [文件: " + event.Event.Message.Content.FileName + "]"
	}

	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.Sender.SenderID,
			ChatID:   event.Event.Chat.ChatID,
			Text:     text,
			Metadata: h.buildMetadata(event),
		},
	}
}

func (h *WebhookHandler) buildInstructionMessage(event *EventMsgVo) *InboundNotification {
	text := event.Event.Message.Content.Text

	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.Sender.SenderID,
			ChatID:   event.Event.Chat.ChatID,
			Text:     text,
			Metadata: h.buildMetadata(event),
		},
	}
}

func (h *WebhookHandler) buildButtonReport(event *EventMsgVo) *InboundNotification {
	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.UserID,
			ChatID:   event.Event.RecvID,
			Text:     event.Event.Value,
			Metadata: map[string]interface{}{
				"peer_kind":  "direct",
				"peer_id":    event.Event.UserID,
				"account_id": h.runtime.AccountID,
			},
		},
	}
}

func (h *WebhookHandler) buildGroupJoin(event *EventMsgVo) *InboundNotification {
	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.UserID,
			ChatID:   event.Event.GroupID,
			Text:     "加入了群聊",
			Metadata: h.buildGroupMetadata(event),
		},
	}
}

func (h *WebhookHandler) buildGroupLeave(event *EventMsgVo) *InboundNotification {
	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.UserID,
			ChatID:   event.Event.GroupID,
			Text:     "退出了群聊",
			Metadata: h.buildGroupMetadata(event),
		},
	}
}

func (h *WebhookHandler) buildBotFollowed(event *EventMsgVo) *InboundNotification {
	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.UserID,
			ChatID:   event.Event.ChatID,
			Text:     "关注了机器人",
			Metadata: map[string]interface{}{
				"peer_kind":  "direct",
				"peer_id":    event.Event.UserID,
				"account_id": h.runtime.AccountID,
			},
		},
	}
}

func (h *WebhookHandler) buildBotUnfollowed(event *EventMsgVo) *InboundNotification {
	return &InboundNotification{
		Message: InboundMessagePayload{
			SenderID: event.Event.UserID,
			ChatID:   event.Event.ChatID,
			Text:     "取消关注了机器人",
			Metadata: map[string]interface{}{
				"peer_kind":  "direct",
				"peer_id":    event.Event.UserID,
				"account_id": h.runtime.AccountID,
			},
		},
	}
}

func (h *WebhookHandler) buildMetadata(event *EventMsgVo) map[string]interface{} {
	peerKind := "direct"
	peerID := event.Event.Sender.SenderID

	chatType := event.Event.Chat.ChatType
	if chatType == "" {
		chatType = event.Event.Message.ChatType
	}

	if chatType == "group" {
		peerKind = "group"
		peerID = event.Event.Chat.ChatID
		if peerID == "" {
			peerID = event.Event.Message.ChatID
		}
	}

	return map[string]interface{}{
		"peer_kind":  peerKind,
		"peer_id":    peerID,
		"account_id": h.runtime.AccountID,
	}
}

func (h *WebhookHandler) buildGroupMetadata(event *EventMsgVo) map[string]interface{} {
	return map[string]interface{}{
		"peer_kind":  "group",
		"peer_id":    event.Event.GroupID,
		"account_id": h.runtime.AccountID,
	}
}

type inboundNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  *InboundNotification   `json:"params"`
}
