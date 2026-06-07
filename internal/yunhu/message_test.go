package yunhu

import (
	"encoding/json"
	"testing"
)

func TestSendMessageRequest_Marshal(t *testing.T) {
	req := SendMessageRequest{
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "hello world",
			Buttons: []Button{
				{Text: "Click", ActionType: ButtonActionReport, Value: "clicked"},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed["recvId"] != "user123" {
		t.Errorf("expected recvId user123")
	}
	if parsed["recvType"] != RecvTypeUser {
		t.Errorf("expected recvType user")
	}
	if parsed["contentType"] != ContentTypeMarkdown {
		t.Errorf("expected contentType markdown")
	}

	content, ok := parsed["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content object")
	}
	if content["text"] != "hello world" {
		t.Errorf("expected text 'hello world', got %v", content["text"])
	}

	buttons, ok := content["buttons"].([]interface{})
	if !ok {
		t.Fatal("expected buttons array")
	}
	if len(buttons) != 1 {
		t.Errorf("expected 1 button, got %d", len(buttons))
	}
}

func TestSendMessageResponse_Unmarshal(t *testing.T) {
	raw := []byte(`{"code":1,"msg":"success","data":{"messageInfo":{"msgId":"msg_xyz","recvId":"user456","recvType":"user"}}}`)
	var resp SendMessageResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 1 {
		t.Errorf("expected code 1, got %d", resp.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil data")
	}
	if resp.Data.MessageInfo == nil {
		t.Fatal("expected non-nil messageInfo")
	}
	if resp.Data.MessageInfo.MsgID != "msg_xyz" {
		t.Errorf("expected msgId msg_xyz, got %s", resp.Data.MessageInfo.MsgID)
	}
}

func TestEditMessageRequest_Marshal(t *testing.T) {
	req := EditMessageRequest{
		MsgID:       "msg_old",
		RecvID:      "user123",
		RecvType:    RecvTypeUser,
		ContentType: ContentTypeMarkdown,
		Content: SendContent{
			Text: "updated text",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["msgId"] != "msg_old" {
		t.Errorf("expected msgId msg_old")
	}
}
