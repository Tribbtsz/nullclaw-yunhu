package channel

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yunhu-channel/yunhu-channel/internal/jsonrpc"
)

type testWriter struct {
	buf strings.Builder
}

func (w *testWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *testWriter) LastOutput() string {
	return w.buf.String()
}

func newTestTransport() (*jsonrpc.StdioTransport, *testWriter) {
	w := &testWriter{}
	t := jsonrpc.NewStdioTransportWithWriter(w)
	return t, w
}

func TestHandler_GetManifest(t *testing.T) {
	transport, _ := newTestTransport()
	h := NewHandler(transport)

	id := json.RawMessage([]byte(`1`))
	req := &jsonrpc.Request{
		JSONRPC: jsonrpc.Version,
		ID:      &id,
		Method:  "get_manifest",
		Params:  json.RawMessage(`{}`),
	}

	h.Handle(nil, req)
}

func TestHandler_MethodNotFound(t *testing.T) {
	transport, w := newTestTransport()
	h := NewHandler(transport)

	id := json.RawMessage([]byte(`99`))
	req := &jsonrpc.Request{
		JSONRPC: jsonrpc.Version,
		ID:      &id,
		Method:  "unknown_method",
		Params:  json.RawMessage(`{}`),
	}

	h.Handle(nil, req)

	output := w.LastOutput()
	if !strings.Contains(output, "method not found") {
		t.Errorf("expected 'method not found' error, got: %s", output)
	}
}

func TestHandler_Health(t *testing.T) {
	transport, _ := newTestTransport()
	h := NewHandler(transport)

	id := json.RawMessage([]byte(`6`))
	req := &jsonrpc.Request{
		JSONRPC: jsonrpc.Version,
		ID:      &id,
		Method:  "health",
		Params:  json.RawMessage(`{}`),
	}

	h.Handle(nil, req)
}

func TestHandler_TrackChatType(t *testing.T) {
	transport, _ := newTestTransport()
	h := NewHandler(transport)

	h.trackChatType([]byte(`{"jsonrpc":"2.0","method":"inbound_message","params":{"message":{"sender_id":"u1","chat_id":"group_001","text":"hi","metadata":{"peer_kind":"group","peer_id":"group_001","account_id":"main"}}}}`))

	recvType := h.inferRecvType("group_001")
	if recvType != "group" {
		t.Errorf("expected recvType group, got %s", recvType)
	}
}

func TestHandler_TrackChatType_Direct(t *testing.T) {
	transport, _ := newTestTransport()
	h := NewHandler(transport)

	h.trackChatType([]byte(`{"jsonrpc":"2.0","method":"inbound_message","params":{"message":{"sender_id":"u1","chat_id":"user_001","text":"hi","metadata":{"peer_kind":"direct","peer_id":"user_001","account_id":"main"}}}}`))

	recvType := h.inferRecvType("user_001")
	if recvType != "user" {
		t.Errorf("expected recvType user, got %s", recvType)
	}
}

func TestHandler_InferRecvType_Unknown(t *testing.T) {
	transport, _ := newTestTransport()
	h := NewHandler(transport)

	recvType := h.inferRecvType("unknown_id")
	if recvType != "user" {
		t.Errorf("expected default recvType user, got %s", recvType)
	}
}
