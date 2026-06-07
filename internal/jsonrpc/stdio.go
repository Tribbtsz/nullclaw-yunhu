package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// ErrParse is returned by ReadRequest when the input is not valid JSON.
var ErrParse = fmt.Errorf("parse error")

// StdioTransport handles reading/writing JSON-RPC messages over stdin/stdout.
type StdioTransport struct {
	reader  *bufio.Scanner
	writer  io.Writer
	writeMu sync.Mutex
}

// NewStdioTransport creates a transport reading from stdin and writing to stdout.
func NewStdioTransport() *StdioTransport {
	s := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 256*1024)
	s.Buffer(buf, 256*1024)
	return &StdioTransport{
		reader: s,
		writer: os.Stdout,
	}
}

// NewStdioTransportWithWriter creates a transport for testing with a custom writer.
func NewStdioTransportWithWriter(w io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewScanner(strings.NewReader("")),
		writer: w,
	}
}

func (t *StdioTransport) ReadRequest() (*Request, error) {
	for t.reader.Scan() {
		line := t.reader.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			slog.Error("failed to parse JSON-RPC request", "error", err)
			return nil, fmt.Errorf("%w: %v", ErrParse, err)
		}
		return &req, nil
	}
	if err := t.reader.Err(); err != nil {
		return nil, fmt.Errorf("stdin read error: %w", err)
	}
	return nil, io.EOF
}

func (t *StdioTransport) WriteResponse(resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	return t.writeLine(data)
}

func (t *StdioTransport) WriteRaw(data []byte) error {
	return t.writeLine(data)
}

func (t *StdioTransport) WriteNotification(method string, params interface{}) error {
	data, err := NewNotification(method, params)
	if err != nil {
		return fmt.Errorf("build notification: %w", err)
	}
	return t.writeLine(data)
}

func (t *StdioTransport) writeLine(data []byte) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if _, err := t.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write to stdout: %w", err)
	}
	return nil
}
