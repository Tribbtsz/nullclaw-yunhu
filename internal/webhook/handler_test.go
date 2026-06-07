package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yunhu-channel/yunhu-channel/internal/config"
)

func TestWebhookHandler_NormalMessage(t *testing.T) {
	var captured []byte
	notifyFn := func(data []byte) {
		captured = data
	}

	cfg := &config.Config{
		WebhookPath: "/webhook/yunhu",
	}
	rt := &config.Runtime{
		Name:      "yunhu",
		AccountID: "main",
	}

	handler := NewWebhookHandler(cfg, rt, notifyFn)

	body := `{
		"version": "1.0",
		"header": {
			"eventId": "evt_001",
			"eventTime": 1700000000000,
			"eventType": "message.receive.normal"
		},
		"event": {
			"sender": {
				"senderId": "user_001",
				"senderType": "user",
				"senderUserLevel": "member",
				"senderNickname": "TestUser"
			},
			"chat": {
				"chatId": "group_001",
				"chatType": "group"
			},
			"message": {
				"msgId": "msg_001",
				"chatType": "group",
				"contentType": "text",
				"content": {
					"text": "Hello World"
				}
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if captured == nil {
		t.Fatal("expected notification to be sent")
	}

	var notif map[string]interface{}
	if err := json.Unmarshal(captured, &notif); err != nil {
		t.Fatal(err)
	}
	if notif["method"] != "inbound_message" {
		t.Errorf("expected method inbound_message, got %v", notif["method"])
	}

	params, ok := notif["params"].(map[string]interface{})
	if !ok {
		t.Fatal("expected params object")
	}
	msg, ok := params["message"].(map[string]interface{})
	if !ok {
		t.Fatal("expected message object")
	}
	if msg["sender_id"] != "user_001" {
		t.Errorf("expected sender_id user_001, got %v", msg["sender_id"])
	}
	if msg["chat_id"] != "group_001" {
		t.Errorf("expected chat_id group_001, got %v", msg["chat_id"])
	}
	if msg["text"] != "Hello World" {
		t.Errorf("expected text 'Hello World', got %v", msg["text"])
	}

	metadata, ok := msg["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected metadata object")
	}
	if metadata["peer_kind"] != "group" {
		t.Errorf("expected peer_kind group, got %v", metadata["peer_kind"])
	}
	if metadata["account_id"] != "main" {
		t.Errorf("expected account_id main, got %v", metadata["account_id"])
	}
}

func TestWebhookHandler_DirectMessage(t *testing.T) {
	var captured []byte
	notifyFn := func(data []byte) {
		captured = data
	}

	cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
	rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

	handler := NewWebhookHandler(cfg, rt, notifyFn)

	body := `{
		"version": "1.0",
		"header": {
			"eventId": "evt_002",
			"eventTime": 1700000000000,
			"eventType": "message.receive.normal"
		},
		"event": {
			"sender": {
				"senderId": "user_002",
				"senderType": "user",
				"senderUserLevel": "member",
				"senderNickname": "DirectUser"
			},
			"chat": {
				"chatId": "user_002",
				"chatType": "bot"
			},
			"message": {
				"msgId": "msg_002",
				"chatType": "bot",
				"contentType": "text",
				"content": {
					"text": "Hi bot"
				}
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if captured == nil {
		t.Fatal("expected notification to be sent")
	}

	var notif map[string]interface{}
	json.Unmarshal(captured, &notif)
	msg := notif["params"].(map[string]interface{})["message"].(map[string]interface{})
	metadata := msg["metadata"].(map[string]interface{})

	if metadata["peer_kind"] != "direct" {
		t.Errorf("expected peer_kind direct, got %v", metadata["peer_kind"])
	}
}

func TestWebhookHandler_ButtonReport(t *testing.T) {
	var captured []byte
	notifyFn := func(data []byte) {
		captured = data
	}

	cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
	rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

	handler := NewWebhookHandler(cfg, rt, notifyFn)

	body := `{
		"version": "1.0",
		"header": {
			"eventId": "evt_003",
			"eventTime": 1700000000000,
			"eventType": "button.report.inline"
		},
		"event": {
			"userId": "user_003",
			"recvId": "group_001",
			"recvType": "group",
			"value": "clicked_option"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if captured == nil {
		t.Fatal("expected notification to be sent")
	}

	var notif map[string]interface{}
	json.Unmarshal(captured, &notif)
	msg := notif["params"].(map[string]interface{})["message"].(map[string]interface{})

	if msg["text"] != "clicked_option" {
		t.Errorf("expected text 'clicked_option', got %v", msg["text"])
	}
	if msg["sender_id"] != "user_003" {
		t.Errorf("expected sender_id user_003, got %v", msg["sender_id"])
	}
}

func TestWebhookHandler_GroupJoin(t *testing.T) {
	var captured []byte
	notifyFn := func(data []byte) {
		captured = data
	}

	cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
	rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

	handler := NewWebhookHandler(cfg, rt, notifyFn)

	body := `{
		"version": "1.0",
		"header": {
			"eventId": "evt_004",
			"eventTime": 1700000000000,
			"eventType": "group.join"
		},
		"event": {
			"userId": "user_004",
			"groupId": "group_002",
			"groupName": "Test Group"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if captured == nil {
		t.Fatal("expected notification to be sent")
	}

	var notif map[string]interface{}
	json.Unmarshal(captured, &notif)
	msg := notif["params"].(map[string]interface{})["message"].(map[string]interface{})

	metadata := msg["metadata"].(map[string]interface{})
	if metadata["peer_kind"] != "group" {
		t.Errorf("expected peer_kind group, got %v", metadata["peer_kind"])
	}
}

func TestWebhookHandler_UnhandledEventType(t *testing.T) {
	var captured []byte
	notifyFn := func(data []byte) {
		captured = data
	}

	cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
	rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

	handler := NewWebhookHandler(cfg, rt, notifyFn)

	body := `{
		"version": "1.0",
		"header": {
			"eventId": "evt_005",
			"eventTime": 1700000000000,
			"eventType": "bot.setting"
		},
		"event": {}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if captured != nil {
		t.Error("expected no notification for unhandled event type")
	}
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
	rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

	handler := NewWebhookHandler(cfg, rt, func([]byte) {})

	req := httptest.NewRequest(http.MethodGet, "/webhook/yunhu", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
