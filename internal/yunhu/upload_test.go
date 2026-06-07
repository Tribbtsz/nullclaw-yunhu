package yunhu

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/image/upload") {
			t.Errorf("expected /image/upload path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "test-token" {
			t.Errorf("expected token test-token, got %s", r.URL.Query().Get("token"))
		}

		resp := UploadResponse{
			Code: 1,
			Msg:  "success",
			Data: &UploadData{
				ImageKey: "img_key_abc",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(testFile, []byte("fake-image-data"), 0644); err != nil {
		t.Fatal(err)
	}

	resp, err := client.Upload("image", testFile)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Code != 1 {
		t.Errorf("expected code 1, got %d", resp.Code)
	}
	if resp.Data == nil || resp.Data.ImageKey != "img_key_abc" {
		t.Errorf("expected imageKey img_key_abc, got %v", resp.Data)
	}
}

func TestUpload_FileNotFound(t *testing.T) {
	client := &Client{
		token:      "test-token",
		httpClient: http.DefaultClient,
	}

	_, err := client.Upload("image", "/nonexistent/file.png")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestUpload_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		resp := UploadResponse{
			Code: -1,
			Msg:  "invalid file",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}
	baseURL = server.URL
	defer func() { baseURL = "https://chat-go.jwzhd.com/open-apis/v1" }()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.png")
	os.WriteFile(testFile, []byte("fake-data"), 0644)

	_, err := client.Upload("image", testFile)
	if err == nil {
		t.Error("expected error for non-1 code")
	}
}
