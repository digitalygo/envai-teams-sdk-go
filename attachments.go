package envai

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"
)

func attachmentMessageIDs(id string, body any) []string {
	if id != "sendChatMessage" {
		return nil
	}
	request, ok := body.(ChatMessageRequest)
	if !ok || !request.AttachmentIDs.Set || request.AttachmentIDs.Null {
		return nil
	}
	return request.AttachmentIDs.Value
}
func sameAttachmentIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (c *Client) UploadAttachmentFile(ctx context.Context, workspaceID, chatID, path, idempotencyKey, contentType string) (ChatAttachment, error) {
	file, err := os.Open(path)
	if err != nil {
		return ChatAttachment{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return ChatAttachment{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxWikiSourceBytes {
		return ChatAttachment{}, errors.New("envai: attachment must be a regular file of at most 1,000,000,000 bytes")
	}
	return c.UploadChatAttachment(ctx, UploadChatAttachmentParams{WorkspaceID: workspaceID, ChatID: chatID, IdempotencyKey: idempotencyKey, Body: []MultipartPart{{Name: "file", Filename: filepath.Base(path), ContentType: contentType, Reader: file, Size: info.Size()}}})
}
func (c *Client) WaitForAttachment(ctx context.Context, workspaceID, chatID, attachmentID string, interval, timeout time.Duration) (ChatAttachment, error) {
	if interval <= 0 || timeout <= 0 {
		return ChatAttachment{}, errors.New("envai: polling bounds must be positive")
	}
	bounded, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		value, err := c.GetChatAttachment(bounded, GetChatAttachmentParams{WorkspaceID: workspaceID, ChatID: chatID, AttachmentID: attachmentID})
		if err != nil {
			return ChatAttachment{}, err
		}
		if value.Lifecycle == "deleting" || value.Lifecycle == "deleted" || value.Conversion.Status == "completed" || value.Conversion.Status == "failed" && !value.Conversion.Recoverable || value.Conversion.Status == "not_required" && value.Lifecycle == "ready" {
			return value, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-bounded.Done():
			timer.Stop()
			return ChatAttachment{}, bounded.Err()
		case <-timer.C:
		}
	}
}
