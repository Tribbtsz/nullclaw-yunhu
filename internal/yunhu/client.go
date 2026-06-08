package yunhu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var (
	baseURL          = "https://chat-go.jwzhd.com/open-apis/v1"
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "yunhu-channel/1.0"
	defaultRetries   = 2
)

type Client struct {
	token      string
	httpClient *http.Client
	maxRetries int
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		maxRetries: defaultRetries,
	}
}

func (c *Client) SendMessage(req *SendMessageRequest) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/bot/send?token=%s", baseURL, c.token)

	resp, err := c.doPost(url, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sendResp SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		return nil, fmt.Errorf("decode send response: %w", err)
	}

	if sendResp.Code != 1 {
		slog.Error("yunhu API send failed",
			"code", sendResp.Code,
			"msg", sendResp.Msg,
			"recvId", req.RecvID,
		)
		return &sendResp, fmt.Errorf("yunhu API error: code=%d msg=%s", sendResp.Code, sendResp.Msg)
	}

	return &sendResp, nil
}

func (c *Client) EditMessage(req *EditMessageRequest) (*EditMessageResponse, error) {
	url := fmt.Sprintf("%s/bot/edit?token=%s", baseURL, c.token)

	resp, err := c.doPost(url, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var editResp EditMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&editResp); err != nil {
		return nil, fmt.Errorf("decode edit response: %w", err)
	}

	if editResp.Code != 1 {
		slog.Error("yunhu API edit failed",
			"code", editResp.Code,
			"msg", editResp.Msg,
			"msgId", req.MsgID,
		)
		return &editResp, fmt.Errorf("yunhu API edit error: code=%d msg=%s", editResp.Code, editResp.Msg)
	}

	return &editResp, nil
}

func (c *Client) RecallMessage(req *RecallMessageRequest) (*RecallMessageResponse, error) {
	url := fmt.Sprintf("%s/bot/recall?token=%s", baseURL, c.token)

	resp, err := c.doPost(url, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var recallResp RecallMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&recallResp); err != nil {
		return nil, fmt.Errorf("decode recall response: %w", err)
	}

	if recallResp.Code != 1 {
		slog.Error("yunhu API recall failed",
			"code", recallResp.Code,
			"msg", recallResp.Msg,
			"msgId", req.MsgID,
		)
		return &recallResp, fmt.Errorf("yunhu API recall error: code=%d msg=%s", recallResp.Code, recallResp.Msg)
	}

	return &recallResp, nil
}

func (c *Client) BatchSend(req *BatchSendRequest) (*BatchSendResponse, error) {
	url := fmt.Sprintf("%s/bot/batch_send?token=%s", baseURL, c.token)

	resp, err := c.doPost(url, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var batchResp BatchSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, fmt.Errorf("decode batch_send response: %w", err)
	}

	if batchResp.Code != 1 {
		slog.Error("yunhu API batch_send failed",
			"code", batchResp.Code,
			"msg", batchResp.Msg,
		)
		return &batchResp, fmt.Errorf("yunhu API batch_send error: code=%d msg=%s", batchResp.Code, batchResp.Msg)
	}

	return &batchResp, nil
}

// StreamWriter 用于向云湖的流式发送通道写入数据。
// 调用 Write 写入文本片段，数据会实时通过 SSE 推送到用户手机。
// 写入完成后调用 Close 关闭连接。
type StreamWriter struct {
	pipeW *io.PipeWriter
	done  chan struct{}
	err   error
}

// Write 写入一段文本到流中，立即推送到用户端。
func (sw *StreamWriter) Write(data []byte) (int, error) {
	return sw.pipeW.Write(data)
}

// Close 完成流式发送，等待云湖 API 响应后返回。
func (sw *StreamWriter) Close() error {
	sw.pipeW.Close()
	<-sw.done
	return sw.err
}

// StartStream 开启一个流式消息发送，返回 *StreamWriter。
// 写入的数据会实时通过 SSE 推送到用户的云湖客户端（类似 ChatGPT 逐字显示效果）。
func (c *Client) StartStream(recvID, recvType, contentType string) (*StreamWriter, error) {
	pipeR, pipeW := io.Pipe()

	url := fmt.Sprintf("%s/bot/send-stream?token=%s&recvId=%s&recvType=%s&contentType=%s",
		baseURL, c.token, recvID, recvType, contentType)

	httpReq, err := http.NewRequest(http.MethodPost, url, pipeR)
	if err != nil {
		pipeW.Close()
		return nil, fmt.Errorf("create stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/plain; charset=utf-8")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	sw := &StreamWriter{
		pipeW: pipeW,
		done:  make(chan struct{}),
	}

	go func() {
		defer close(sw.done)
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in stream request", "recover", r)
			}
		}()

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			sw.err = fmt.Errorf("stream request: %w", err)
			return
		}
		defer resp.Body.Close()

		var streamResp SendStreamResponse
		if err := json.NewDecoder(resp.Body).Decode(&streamResp); err != nil {
			sw.err = fmt.Errorf("decode stream response: %w", err)
			return
		}

		if streamResp.Code != 1 {
			sw.err = fmt.Errorf("yunhu API send-stream error: code=%d msg=%s", streamResp.Code, streamResp.Msg)
			return
		}
	}()

	return sw, nil
}

func (c *Client) Ping() error {
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("ping returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) doPost(url string, body interface{}) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			slog.Warn("retrying yunhu API request", "attempt", attempt, "backoff", backoff)
			time.Sleep(backoff)
		}

		resp, err := c.doPostOnce(url, body)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("retry exhausted: %w", lastErr)
}

func (c *Client) doPostOnce(url string, body interface{}) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("yunhu API returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func isRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "status 5") ||
		strings.Contains(msg, "status 429")
}
