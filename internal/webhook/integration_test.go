package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunhu-channel/yunhu-channel/internal/config"
)

func TestWebhookHandler_Integration_AllEvents(t *testing.T) {
	tests := []struct {
		filename       string
		expectNotif    bool
		expectSenderID string
		expectChatID   string
		expectPeerKind string
		expectText     string
	}{
		{
			filename:       "webhook_normal_message.json",
			expectNotif:    true,
			expectSenderID: "user_001",
			expectChatID:   "group_001",
			expectPeerKind: "group",
			expectText:     "你好，这是一条测试消息",
		},
		{
			filename:       "webhook_instruction_message.json",
			expectNotif:    true,
			expectSenderID: "user_002",
			expectChatID:   "user_002",
			expectPeerKind: "direct",
			expectText:     "/查询 天气",
		},
		{
			filename:       "webhook_button_report.json",
			expectNotif:    true,
			expectSenderID: "user_003",
			expectChatID:   "group_001",
			expectPeerKind: "direct",
			expectText:     "opt1",
		},
		{
			filename:       "webhook_group_join.json",
			expectNotif:    true,
			expectSenderID: "user_004",
			expectChatID:   "group_002",
			expectPeerKind: "group",
			expectText:     "加入了群聊",
		},
		{
			filename:       "webhook_bot_followed.json",
			expectNotif:    true,
			expectSenderID: "user_005",
			expectChatID:   "user_005",
			expectPeerKind: "direct",
			expectText:     "关注了机器人",
		},
		{
			filename:       "webhook_bot_unfollowed.json",
			expectNotif:    true,
			expectSenderID: "user_005",
			expectChatID:   "user_005",
			expectPeerKind: "direct",
			expectText:     "取消关注了机器人",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			var captured []byte
			notifyFn := func(data []byte) {
				captured = data
			}

			cfg := &config.Config{WebhookPath: "/webhook/yunhu"}
			rt := &config.Runtime{Name: "yunhu", AccountID: "main"}

			handler := NewWebhookHandler(cfg, rt, notifyFn)

			body := loadTestData(t, tt.filename)
			req := httptest.NewRequest(http.MethodPost, "/webhook/yunhu", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			if tt.expectNotif && captured == nil {
				t.Fatal("expected notification but got nil")
			}
			if !tt.expectNotif && captured != nil {
				t.Fatal("expected no notification but got one")
			}

			if captured == nil {
				return
			}

			var notif map[string]interface{}
			if err := json.Unmarshal(captured, &notif); err != nil {
				t.Fatalf("invalid notification JSON: %v", err)
			}

			params := notif["params"].(map[string]interface{})
			msg := params["message"].(map[string]interface{})

			if msg["sender_id"] != tt.expectSenderID {
				t.Errorf("sender_id: expected %s, got %v", tt.expectSenderID, msg["sender_id"])
			}
			if msg["chat_id"] != tt.expectChatID {
				t.Errorf("chat_id: expected %s, got %v", tt.expectChatID, msg["chat_id"])
			}
			if msg["text"] != tt.expectText {
				t.Errorf("text: expected %s, got %v", tt.expectText, msg["text"])
			}

			metadata := msg["metadata"].(map[string]interface{})
			if metadata["peer_kind"] != tt.expectPeerKind {
				t.Errorf("peer_kind: expected %s, got %v", tt.expectPeerKind, metadata["peer_kind"])
			}
			if metadata["account_id"] != "main" {
				t.Errorf("account_id: expected main, got %v", metadata["account_id"])
			}
		})
	}
}

func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", filename, err)
	}
	return data
}
