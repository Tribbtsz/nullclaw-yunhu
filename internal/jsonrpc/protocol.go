package jsonrpc

import (
	"encoding/json"
	"fmt"
)

const Version = "2.0"

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (r *Request) IsNotification() bool {
	return r.ID == nil
}

func (r *Request) HasID() bool {
	return r.ID != nil
}

type Response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInternalError  = -32603
)
