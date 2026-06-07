package yunhu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var (
	baseURL          = "https://chat-go.jwzhd.com/open-apis/v1"
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "yunhu-channel/1.0"
)

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
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

func (c *Client) Ping() error {
	url := fmt.Sprintf("%s/bot/send?token=%s", baseURL, c.token)
	pingReq := &SendMessageRequest{
		RecvID:      "__ping__",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeText,
		Content: SendContent{
			Text: "",
		},
	}

	body, err := json.Marshal(pingReq)
	if err != nil {
		return fmt.Errorf("marshal ping: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create ping request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("ping returned status %d", resp.StatusCode)
}

func (c *Client) doPost(url string, body interface{}) (*http.Response, error) {
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
