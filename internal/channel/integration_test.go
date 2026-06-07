package channel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunhu-channel/yunhu-channel/internal/jsonrpc"
)

func TestHandler_Integration_GetManifest(t *testing.T) {
	transport, w := newTestTransport()
	h := NewHandler(transport)

	req := loadTestRequest(t, "get_manifest_request.json")
	expectedResp := loadTestFile(t, "get_manifest_response.json")

	h.Handle(nil, req)

	output := strings.TrimSpace(w.LastOutput())
	var actual map[string]interface{}
	if err := json.Unmarshal([]byte(output), &actual); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}

	var expected map[string]interface{}
	if err := json.Unmarshal(expectedResp, &expected); err != nil {
		t.Fatalf("invalid expected JSON: %v", err)
	}

	actualResult := actual["result"].(map[string]interface{})
	expectedResult := expected["result"].(map[string]interface{})

	if int(actualResult["protocol_version"].(float64)) != 2 {
		t.Error("protocol_version should be 2")
	}

	actualCaps := actualResult["capabilities"].(map[string]interface{})
	expectedCaps := expectedResult["capabilities"].(map[string]interface{})

	for k, v := range expectedCaps {
		if actualCaps[k] != v {
			t.Errorf("capability %s: expected %v, got %v", k, v, actualCaps[k])
		}
	}
}

func TestHandler_Integration_MethodNotFound(t *testing.T) {
	transport, w := newTestTransport()
	h := NewHandler(transport)

	id := json.RawMessage([]byte(`99`))
	req := &jsonrpc.Request{
		JSONRPC: jsonrpc.Version,
		ID:      &id,
		Method:  "nonexistent_method",
		Params:  json.RawMessage(`{}`),
	}

	h.Handle(nil, req)

	output := w.LastOutput()
	if !strings.Contains(output, "method not found") {
		t.Errorf("expected 'method not found' error, got: %s", output)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &resp); err != nil {
		t.Fatal(err)
	}
	errObj := resp["error"].(map[string]interface{})
	if int(errObj["code"].(float64)) != jsonrpc.CodeMethodNotFound {
		t.Errorf("expected error code %d, got %v", jsonrpc.CodeMethodNotFound, errObj["code"])
	}
}

func TestHandler_Integration_ParseError(t *testing.T) {
	transport := jsonrpc.NewStdioTransportWithWriter(nil)
	h := NewHandler(transport)

	h.Handle(nil, nil)
}

func TestHandler_Integration_StopNotStarted(t *testing.T) {
	transport, w := newTestTransport()
	h := NewHandler(transport)

	id := json.RawMessage([]byte(`3`))
	stopReq := &jsonrpc.Request{
		JSONRPC: jsonrpc.Version,
		ID:      &id,
		Method:  "stop",
		Params:  json.RawMessage(`{}`),
	}

	h.Handle(nil, stopReq)

	output := strings.TrimSpace(w.LastOutput())
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatal(err)
	}
	result := resp["result"].(map[string]interface{})
	if result["accepted"] != true {
		t.Errorf("expected accepted:true, got %v", result["accepted"])
	}
}

func loadTestRequest(t *testing.T, filename string) *jsonrpc.Request {
	t.Helper()
	data := loadTestFile(t, filename)
	var req jsonrpc.Request
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("failed to parse %s: %v", filename, err)
	}
	return &req
}

func loadTestFile(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return data
}
