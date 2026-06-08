package channel

import (
	"encoding/json"
	"testing"

	jsonrpcutil "github.com/yunhu-channel/yunhu-channel/internal/jsonrpc"
)

func TestBuildManifest(t *testing.T) {
	m := BuildManifest()
	if m.ProtocolVersion != 2 {
		t.Errorf("expected protocol version 2, got %d", m.ProtocolVersion)
	}
	if !m.Capabilities.Health {
		t.Error("expected health capability")
	}
	if !m.Capabilities.Streaming {
		t.Error("expected streaming capability")
	}
	if !m.Capabilities.SendRich {
		t.Error("expected send_rich capability")
	}
	if !m.Capabilities.Edit {
		t.Error("expected edit capability")
	}
	if !m.Capabilities.Typing {
		t.Error("typing should be true")
	}
	if !m.Capabilities.Delete {
		t.Error("expected delete capability")
	}
	if m.Capabilities.Reactions {
		t.Error("reactions should be false")
	}
	if m.Capabilities.ReadReceipts {
		t.Error("read_receipts should be false")
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	caps, ok := parsed["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("expected capabilities object")
	}
	if caps["health"] != true {
		t.Error("expected health: true")
	}
	if caps["edit"] != true {
		t.Error("expected edit: true")
	}
}

func TestManifestResponseFormat(t *testing.T) {
	m := BuildManifest()
	id := json.RawMessage([]byte(`1`))
	resp, err := jsonrpcutil.NewResponse(&id, m)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	result, ok := parsed["result"].(map[string]interface{})
	if !ok {
		t.Fatal("expected result object")
	}
	if int(result["protocol_version"].(float64)) != 2 {
		t.Error("expected protocol_version 2")
	}
}

func TestStartResult(t *testing.T) {
	result := StartResult{Started: true}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"started":true}` {
		t.Errorf("unexpected json: %s", string(data))
	}
}

func TestStopResult(t *testing.T) {
	result := StopResult{Accepted: true}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"accepted":true}` {
		t.Errorf("unexpected json: %s", string(data))
	}
}

func TestSendResult(t *testing.T) {
	result := SendResult{Accepted: true, MessageID: "msg_123"}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"accepted":true,"message_id":"msg_123"}` {
		t.Errorf("unexpected json: %s", string(data))
	}
}

func TestHealthResult(t *testing.T) {
	result := HealthResult{Healthy: true}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"healthy":true}` {
		t.Errorf("unexpected json: %s", string(data))
	}
}

func TestDeleteResult(t *testing.T) {
	result := DeleteResult{Accepted: true}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"accepted":true}` {
		t.Errorf("unexpected json: %s", string(data))
	}
}
