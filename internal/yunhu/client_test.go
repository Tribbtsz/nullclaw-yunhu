package yunhu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("token") != "test-token" {
			t.Errorf("expected token test-token, got %s", r.URL.Query().Get("token"))
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.RecvID != "user123" {
			t.Errorf("expected recvId user123, got %s", req.RecvID)
		}
		if req.ContentType != ContentTypeMarkdown {
			t.Errorf("expected contentType markdown, got %s", req.ContentType)
		}
		if req.Content.Text != "hello" {
			t.Errorf("expected text hello, got %s", req.Content.Text)
		}

		resp := SendMessageResponse{
			Code: 1,
			Msg:  "success",
			Data: &SendMessageData{
				MessageInfo: &MessageInfo{
					MsgID:    "msg_abc123",
					RecvID:   "user123",
					RecvType: RecvTypeUser,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	resp, err := client.SendMessage(&SendMessageRequest{
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "hello",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Code != 1 {
		t.Errorf("expected code 1, got %d", resp.Code)
	}
	if resp.Data.MessageInfo.MsgID != "msg_abc123" {
		t.Errorf("expected msgId msg_abc123, got %s", resp.Data.MessageInfo.MsgID)
	}
}

func TestSendMessage_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SendMessageResponse{
			Code: -1,
			Msg:  "token invalid",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "bad-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	_, err := client.SendMessage(&SendMessageRequest{
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "hello",
		},
	})
	if err == nil {
		t.Error("expected error for non-1 code")
	}
}

func TestEditMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !stringsContains(r.URL.Path, "/bot/edit") {
			t.Errorf("expected /bot/edit path, got %s", r.URL.Path)
		}

		var req EditMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.MsgID != "msg_old" {
			t.Errorf("expected msgId msg_old, got %s", req.MsgID)
		}
		if req.Content.Text != "edited text" {
			t.Errorf("expected text 'edited text', got %s", req.Content.Text)
		}

		resp := EditMessageResponse{
			Code: 1,
			Msg:  "success",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	_, err := client.EditMessage(&EditMessageRequest{
		MsgID:       "msg_old",
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "edited text",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendMessage_WithButtons(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if len(req.Content.Buttons) != 2 {
			t.Errorf("expected 2 buttons, got %d", len(req.Content.Buttons))
		}
		if req.Content.Buttons[0].ActionType != ButtonActionReport {
			t.Errorf("expected actionType 3, got %d", req.Content.Buttons[0].ActionType)
		}
		if req.Content.Buttons[0].Value != "opt1" {
			t.Errorf("expected value opt1, got %s", req.Content.Buttons[0].Value)
		}

		resp := SendMessageResponse{Code: 1, Msg: "success"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	_, err := client.SendMessage(&SendMessageRequest{
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "choose",
			Buttons: []Button{
				{Text: "Opt1", ActionType: ButtonActionReport, Value: "opt1"},
				{Text: "Opt2", ActionType: ButtonActionReport, Value: "opt2"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
