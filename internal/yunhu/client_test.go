package yunhu

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		if !strings.Contains(r.URL.Path, "/bot/edit") {
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

func TestRecallMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RecallMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.MsgID != "msg_to_delete" {
			t.Errorf("expected msgId msg_to_delete, got %s", req.MsgID)
		}
		if req.ChatType != "user" {
			t.Errorf("expected chatType user, got %s", req.ChatType)
		}

		resp := RecallMessageResponse{Code: 1, Msg: "success"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	_, err := client.RecallMessage(&RecallMessageRequest{
		MsgID:    "msg_to_delete",
		ChatType: "user",
		ChatID:   "user123",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBatchSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BatchSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if len(req.RecvIDs) != 2 {
			t.Errorf("expected 2 recvIds, got %d", len(req.RecvIDs))
		}
		if req.Content.Text != "batch hello" {
			t.Errorf("expected text 'batch hello', got %s", req.Content.Text)
		}

		resp := BatchSendResponse{
			Code: 1,
			Msg:  "success",
			Data: &BatchSendData{
				SuccessCount: "2",
				SuccessList: []BatchSendMsgInfo{
					{MsgID: "m1", RecvID: "u1", RecvType: "user"},
					{MsgID: "m2", RecvID: "u2", RecvType: "user"},
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

	resp, err := client.BatchSend(&BatchSendRequest{
		RecvIDs:     []string{"u1", "u2"},
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "batch hello",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.SuccessCount != "2" {
		t.Errorf("expected successCount 2, got %s", resp.Data.SuccessCount)
	}
}

func TestStartStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("recvId") != "user123" {
			t.Errorf("expected recvId user123, got %s", r.URL.Query().Get("recvId"))
		}
		if r.URL.Query().Get("recvType") != "user" {
			t.Errorf("expected recvType user, got %s", r.URL.Query().Get("recvType"))
		}
		if r.URL.Query().Get("contentType") != "markdown" {
			t.Errorf("expected contentType markdown, got %s", r.URL.Query().Get("contentType"))
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Errorf("expected body 'hello', got %q", string(body))
		}

		resp := SendStreamResponse{Code: 1, Msg: "success"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	sw, err := client.StartStream("user123", RecvTypeUser, ContentTypeMarkdown)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sw.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
}
