package envai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// MaxWikiSourceBytes is the inclusive durable original upload limit.
const MaxWikiSourceBytes int64 = 1_000_000_000

// UploadSourceFile streams an unchanged local regular file. Replays are explicit:
// reuse the original bytes and idempotency key after reconciling ambiguous results.
func (c *Client) UploadSourceFile(ctx context.Context, workspaceID, path, idempotencyKey, contentType string) (WikiSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return WikiSource{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return WikiSource{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxWikiSourceBytes {
		return WikiSource{}, errors.New("envai: source must be a regular file of at most 1,000,000,000 bytes")
	}
	return c.UploadWikiSource(ctx, UploadWikiSourceParams{WorkspaceID: workspaceID, IdempotencyKey: idempotencyKey, Body: []MultipartPart{{Name: "file", Filename: filepath.Base(path), ContentType: contentType, Reader: f, Size: info.Size()}}})
}
func (c *Client) validateOperationParams(op operation, params map[string]any) error {
	for _, p := range op.Parameters {
		if p.In == "header" {
			if value, exists := params[p.Name]; exists {
				v := fmt.Sprint(value)
				if len(v) < 1 || len(v) > 255 {
					return errors.New("envai: invalid header parameter")
				}
				for _, r := range v {
					if r < 33 || r > 126 {
						return errors.New("envai: invalid header parameter")
					}
				}
			}
		}
	}
	if op.ID == "getWikiCategoryIndex" {
		from, f := params["from"]
		to, t := params["to"]
		all, a := params["all"]
		if a && all == true {
			if f || t {
				return errors.New("envai: all cannot be combined with range")
			}
		} else if !f || !t {
			return errors.New("envai: select all or an inclusive from/to range")
		}
		if f && (from.(int64) < 1 || to.(int64) < from.(int64)) {
			return errors.New("envai: invalid inclusive range")
		}
		if typ, ok := params["type"]; ok && typ != "raw" && typ != "artifact" {
			return errors.New("envai: invalid catalog entry type")
		}
		if rev, ok := params["catalog_revision"]; ok && rev.(int64) < 0 {
			return errors.New("envai: invalid catalog revision")
		}
	}
	return nil
}
func validateWikiBody(op operation, body any) error {
	if op.ID == "uploadWikiSource" {
		parts, ok := body.([]MultipartPart)
		if !ok || len(parts) != 1 || parts[0].Name != "file" || parts[0].Filename == "" {
			return errors.New("envai: exactly one named source file is required")
		}
		p := parts[0]
		size := int64(len(p.Data))
		if p.Reader != nil {
			if p.Data != nil {
				return errors.New("envai: use source reader or bytes, not both")
			}
			size = p.Size
		}
		if size < 0 || size > MaxWikiSourceBytes {
			return errors.New("envai: invalid source byte size")
		}
	}
	var webhook string
	switch v := body.(type) {
	case CreateWikiInboxRunBody:
		webhook = v.WebhookURL
	case CreateWikiOutboxRunBody:
		webhook = v.WebhookURL
	case CreateWikiSyncRunBody:
		webhook = v.WebhookURL
	case WithdrawWikiSourceBody:
		if v.FileID == "" {
			return errors.New("envai: source file ID is required")
		}
	}
	if strings.HasPrefix(op.ID, "createWiki") {
		u, err := url.Parse(webhook)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
			return errors.New("envai: webhook requires an HTTPS URL without credentials or fragment")
		}
	}
	return nil
}

// multipartReader constructs only small framing buffers. Original bytes are read
// lazily by net/http, and no request replay body is installed for mutations.
func multipartReader(parts []MultipartPart) (io.Reader, string, error) {
	var framing bytes.Buffer
	w := multipart.NewWriter(&framing)
	readers := make([]io.Reader, 0, len(parts)*2+1)
	offset := 0
	for _, p := range parts {
		if p.Name == "" || strings.ContainsAny(p.Name+p.Filename+p.ContentType, "\r\n\x00") {
			return nil, "", errors.New("envai: invalid multipart metadata")
		}
		h := textproto.MIMEHeader{}
		escape := strings.NewReplacer("\\", "\\\\", "\"", "\\\"")
		disposition := `form-data; name="` + escape.Replace(p.Name) + `"`
		if p.Filename != "" {
			disposition += `; filename="` + escape.Replace(p.Filename) + `"`
		}
		h.Set("Content-Disposition", disposition)
		if p.ContentType != "" {
			h.Set("Content-Type", p.ContentType)
		}
		if _, err := w.CreatePart(h); err != nil {
			return nil, "", err
		}
		header := append([]byte(nil), framing.Bytes()[offset:]...)
		offset = framing.Len()
		readers = append(readers, bytes.NewReader(header))
		if p.Reader != nil {
			if p.Size < 0 {
				return nil, "", errors.New("envai: multipart reader requires a nonnegative size")
			}
			readers = append(readers, &exactReader{reader: p.Reader, remaining: p.Size})
		} else {
			readers = append(readers, bytes.NewReader(p.Data))
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	readers = append(readers, bytes.NewReader(append([]byte(nil), framing.Bytes()[offset:]...)))
	return io.MultiReader(readers...), w.FormDataContentType(), nil
}

type exactReader struct {
	reader    io.Reader
	remaining int64
}

func (r *exactReader) Read(b []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	if int64(len(b)) > r.remaining {
		b = b[:r.remaining]
	}
	n, err := r.reader.Read(b)
	r.remaining -= int64(n)
	if err == io.EOF && r.remaining > 0 {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}
