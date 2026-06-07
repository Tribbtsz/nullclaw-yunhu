package jsonrpc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStdioTransport_ReadRequest(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"get_manifest","params":{}}
{"jsonrpc":"2.0","method":"inbound_message","params":{"message":{"sender_id":"u1","chat_id":"c1","text":"hi"}}}
`

	reader := strings.NewReader(input)
	transport := &StdioTransport{}

	originalReader := transport.reader
	defer func() { transport.reader = originalReader }()

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"get_manifest","params":{}}`,
		`{"jsonrpc":"2.0","method":"inbound_message","params":{"message":{"sender_id":"u1","chat_id":"c1","text":"hi"}}}`,
	}
	for i, expectedLine := range lines {
		var req Request
		if err := json.Unmarshal([]byte(expectedLine), &req); err != nil {
			t.Fatalf("line %d: parse error: %v", i, err)
		}
		if i == 0 && !req.HasID() {
			t.Error("request 0 should have id")
		}
		if i == 1 && !req.IsNotification() {
			t.Error("request 1 should be notification")
		}
	}
	_ = reader
}

func TestStdioTransport_WriteResponse(t *testing.T) {
	var buf strings.Builder
	transport := &StdioTransport{
		writer: &buf,
	}

	id := json.RawMessage([]byte(`1`))
	resp := NewErrorResponse(&id, CodeMethodNotFound, "not found")
	if err := transport.WriteResponse(resp); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}

	var parsed Response
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Error == nil {
		t.Fatal("expected error in response")
	}
	if parsed.Error.Code != CodeMethodNotFound {
		t.Errorf("expected code %d, got %d", CodeMethodNotFound, parsed.Error.Code)
	}
}

func TestStdioTransport_WriteNotification(t *testing.T) {
	var buf strings.Builder
	transport := &StdioTransport{
		writer: &buf,
	}

	params := map[string]interface{}{
		"message": map[string]interface{}{
			"sender_id": "user1",
			"chat_id":   "room1",
			"text":      "hello",
		},
	}
	if err := transport.WriteNotification("inbound_message", params); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}

	var req Request
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &req); err != nil {
		t.Fatal(err)
	}
	if req.Method != "inbound_message" {
		t.Errorf("expected inbound_message, got %s", req.Method)
	}
	if req.ID != nil {
		t.Error("notification should not have id")
	}
}

func TestStdioTransport_WriteRaw(t *testing.T) {
	var buf strings.Builder
	transport := &StdioTransport{
		writer: &buf,
	}

	raw := []byte(`{"jsonrpc":"2.0","method":"inbound_message","params":{"message":{"sender_id":"u1"}}}`)
	if err := transport.WriteRaw(raw); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}
}
