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

// Upload 将本地文件上传到云湖，返回文件 key。
//
// uploadType 可选值: "image" / "video" / "file"
//
// 拿到 key 后通过发送消息接口的 imageKey/videoKey/fileKey 字段发送：
//
//	client.Upload("image", "/path/to/photo.png")
//	// 返回的 UploadResponse.Data.ImageKey 用于发送消息
//	client.SendMessage(&SendMessageRequest{
//	    RecvID:      target,
//	    RecvType:    recvType,
//	    ContentType: ContentTypeImage,
//	    Content:     SendContent{ImageKey: imageKey},
//	})
//
// 注意：不能直接传公开 URL 给 imageKey/videoKey/fileKey，
// 必须通过此接口上传拿到 key 后再发送。
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
