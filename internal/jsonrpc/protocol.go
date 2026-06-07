package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Version is the JSON-RPC protocol version.
const Version = "2.0"

// CodeParseError is the JSON-RPC error code for invalid JSON.
// CodeInvalidRequest is the JSON-RPC error code for invalid request structure.
// CodeMethodNotFound is the JSON-RPC error code for unknown method names.
// CodeInternalError is the JSON-RPC error code for internal execution errors.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInternalError  = -32603
)

// Request represents a JSON-RPC 2.0 request or notification.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether this request has no id field.
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// HasID reports whether this request has an id field.
func (r *Request) HasID() bool {
	return r.ID != nil
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewResponse creates a successful JSON-RPC response.
func NewResponse(id *json.RawMessage, result interface{}) (*Response, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	rm := json.RawMessage(raw)
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Result:  rm,
	}, nil
}

// NewErrorResponse creates a JSON-RPC error response.
func NewErrorResponse(id *json.RawMessage, code int, message string) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
}

// NewNotification creates a JSON-RPC notification (no id field).
func NewNotification(method string, params interface{}) ([]byte, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}
	req := &Request{
		JSONRPC: Version,
		Method:  method,
		Params:  json.RawMessage(raw),
	}
	return json.Marshal(req)
}
