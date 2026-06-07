package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestNewResponse(t *testing.T) {
	id := json.RawMessage([]byte(`1`))
	resp, err := NewResponse(&id, map[string]bool{"started": true})
	if err != nil {
		t.Fatal(err)
	}
	if resp.JSONRPC != Version {
		t.Errorf("expected jsonrpc %s, got %s", Version, resp.JSONRPC)
	}
	if resp.ID == nil {
		t.Fatal("expected non-nil ID")
	}
	if resp.Error != nil {
		t.Error("expected nil error")
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := json.RawMessage([]byte(`2`))
	resp := NewErrorResponse(&id, CodeMethodNotFound, "method not found: foo")
	if resp.JSONRPC != Version {
		t.Errorf("expected jsonrpc %s, got %s", Version, resp.JSONRPC)
	}
	if resp.Result != nil {
		t.Error("expected nil result")
	}
	if resp.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("expected code %d, got %d", CodeMethodNotFound, resp.Error.Code)
	}
}

func TestNewNotification(t *testing.T) {
	params := map[string]interface{}{
		"message": map[string]interface{}{
			"sender_id": "user1",
			"chat_id":   "room1",
			"text":      "hello",
		},
	}
	data, err := NewNotification("inbound_message", params)
	if err != nil {
		t.Fatal(err)
	}
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatal(err)
	}
	if req.Method != "inbound_message" {
		t.Errorf("expected method inbound_message, got %s", req.Method)
	}
	if req.ID != nil {
		t.Error("notification should not have id")
	}
}

func TestRequestIsNotification(t *testing.T) {
	req := &Request{JSONRPC: Version, Method: "inbound_message"}
	if !req.IsNotification() {
		t.Error("expected IsNotification to return true")
	}
	if req.HasID() {
		t.Error("expected HasID to return false")
	}
}

func TestRequestHasID(t *testing.T) {
	id := json.RawMessage([]byte(`3`))
	req := &Request{JSONRPC: Version, ID: &id, Method: "send"}
	if req.IsNotification() {
		t.Error("expected IsNotification to return false")
	}
	if !req.HasID() {
		t.Error("expected HasID to return true")
	}
}

func TestMarshalUnmarshalRequest(t *testing.T) {
	raw := []byte(`{"jsonrpc":"2.0","id":1,"method":"get_manifest","params":{}}`)
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	if req.Method != "get_manifest" {
		t.Errorf("expected get_manifest, got %s", req.Method)
	}
	if req.ID == nil {
		t.Fatal("expected non-nil ID")
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty marshal output")
	}
}

func TestMarshalUnmarshalResponse(t *testing.T) {
	id := json.RawMessage([]byte(`7`))
	resp, err := NewResponse(&id, map[string]bool{"healthy": true})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var parsed Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.ID == nil {
		t.Fatal("expected non-nil ID")
	}
}
