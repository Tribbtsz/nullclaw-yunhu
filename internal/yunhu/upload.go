package yunhu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type UploadResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *UploadData `json:"data,omitempty"`
}

type UploadData struct {
	ImageKey string `json:"imageKey,omitempty"`
	VideoKey string `json:"videoKey,omitempty"`
	FileKey  string `json:"fileKey,omitempty"`
}

func (c *Client) Upload(uploadType string, filePath string) (*UploadResponse, error) {
	url := fmt.Sprintf("%s/%s/upload?token=%s", baseURL, uploadType, c.token)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(uploadType, filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	writer.Close()

	httpReq, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create upload request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("decode upload response: %w", err)
	}

	if uploadResp.Code != 1 {
		return &uploadResp, fmt.Errorf("yunhu upload error: code=%d msg=%s", uploadResp.Code, uploadResp.Msg)
	}

	return &uploadResp, nil
}
