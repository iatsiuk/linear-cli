package api

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const fileUploadMutation = `
mutation FileUpload($contentType: String!, $filename: String!, $size: Int!) {
	fileUpload(contentType: $contentType, filename: $filename, size: $size) {
		success
		uploadFile {
			assetUrl
			uploadUrl
			headers {
				key
				value
			}
		}
	}
}
`

type uploadFileHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type uploadFile struct {
	AssetUrl  string             `json:"assetUrl"`
	UploadUrl string             `json:"uploadUrl"`
	Headers   []uploadFileHeader `json:"headers"`
}

type uploadPayload struct {
	FileUpload struct {
		Success    bool        `json:"success"`
		UploadFile *uploadFile `json:"uploadFile"`
	} `json:"fileUpload"`
}

// Upload uploads a file to Linear's storage and returns the assetUrl.
// It calls the fileUpload mutation to get a signed upload URL, then
// PUTs the file to that URL.
func (c *Client) Upload(ctx context.Context, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	filename := filepath.Base(filePath)
	contentType := contentTypeFromName(filename)
	if info.Size() > math.MaxInt32 {
		return "", fmt.Errorf("file too large: %d bytes (max 2 GB)", info.Size())
	}
	size := int(info.Size())

	vars := map[string]any{
		"contentType": contentType,
		"filename":    filename,
		"size":        size,
	}

	var payload uploadPayload
	if err := c.Do(ctx, fileUploadMutation, vars, &payload); err != nil {
		return "", fmt.Errorf("fileUpload mutation: %w", err)
	}
	if !payload.FileUpload.Success || payload.FileUpload.UploadFile == nil {
		return "", fmt.Errorf("fileUpload mutation: no upload file in response")
	}

	uf := payload.FileUpload.UploadFile

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uf.UploadUrl, f)
	if err != nil {
		return "", fmt.Errorf("create PUT request: %w", err)
	}
	req.ContentLength = int64(size)
	req.Header.Set("Content-Type", contentType)
	for _, h := range uf.Headers {
		req.Header.Set(h.Key, h.Value)
	}

	// use a client without timeout for the upload PUT - file uploads can take longer than API calls
	uploadClient := &http.Client{}
	resp, err := uploadClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload file: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload file: unexpected status %d", resp.StatusCode)
	}

	return uf.AssetUrl, nil
}

func contentTypeFromName(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}
	t := mime.TypeByExtension(ext)
	if t == "" {
		return "application/octet-stream"
	}
	// strip parameters like "; charset=utf-8" that some platforms (macOS) append
	if i := strings.IndexByte(t, ';'); i >= 0 {
		t = strings.TrimSpace(t[:i])
	}
	return t
}
