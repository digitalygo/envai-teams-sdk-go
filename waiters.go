package envai

import (
	"context"
	"errors"
	"time"
)

// PollOptions bounds the whole operation, including admission when applicable.
type PollOptions struct{ Interval, Timeout time.Duration }

func (o PollOptions) bounded(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if o.Interval <= 0 || o.Timeout <= 0 {
		return nil, nil, errors.New("envai: polling bounds must be positive")
	}
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	return ctx, cancel, nil
}

// WaitForSource only observes conversion; it never starts indexing or retries work.
func (c *Client) WaitForSource(ctx context.Context, workspaceID, fileID string, options PollOptions) (WikiSource, error) {
	ctx, cancel, err := options.bounded(ctx)
	if err != nil {
		return WikiSource{}, err
	}
	defer cancel()
	for {
		source, err := c.GetWikiSource(ctx, GetWikiSourceParams{WorkspaceID: workspaceID, FileID: fileID})
		if err != nil {
			return source, err
		}
		if source.WorkspaceID != workspaceID || source.FileID != fileID {
			return source, &ProtocolError{"source identity mismatch"}
		}
		terminal := source.Conversion.Status == "completed" || source.Conversion.Status == "not_required" && (source.Lifecycle == "inbox" || source.Lifecycle == "raw") || source.Conversion.Status == "failed" && !source.Conversion.Recoverable.Value || source.Lifecycle == "deleted" || source.Lifecycle == "deleting" || source.Lifecycle == "outbox"
		if terminal {
			return source, ctx.Err()
		}
		if err = wait(ctx, options.Interval); err != nil {
			return source, err
		}
	}
}
func (c *Client) WaitForWikiOperation(ctx context.Context, workspaceID, operationID string, options PollOptions) (WikiOperation, error) {
	ctx, cancel, err := options.bounded(ctx)
	if err != nil {
		return WikiOperation{}, err
	}
	defer cancel()
	for {
		operation, err := c.GetWikiOperation(ctx, GetWikiOperationParams{WorkspaceID: workspaceID, OperationID: operationID})
		if err != nil {
			return operation, err
		}
		if operation.WorkspaceID != workspaceID || operation.OperationID != operationID {
			return operation, &ProtocolError{"wiki operation identity mismatch"}
		}
		if operation.Status == "completed" || operation.Status == "failed" {
			return operation, ctx.Err()
		}
		if err = wait(ctx, options.Interval); err != nil {
			return operation, err
		}
	}
}

// GenerateAndWait admits exactly once and then observes the record. It never
// follows status_url and never resubmits an uncertain mutation.
func (c *Client) GenerateAndWait(ctx context.Context, request CreateImageGenerationParams, options PollOptions) (ImageGenerationStatusRecord, error) {
	ctx, cancel, err := options.bounded(ctx)
	if err != nil {
		return ImageGenerationStatusRecord{}, err
	}
	defer cancel()
	accepted, err := c.CreateImageGeneration(ctx, request)
	if err != nil {
		return ImageGenerationStatusRecord{}, err
	}
	if accepted.ID == "" {
		return ImageGenerationStatusRecord{}, &ProtocolError{"missing image record identity"}
	}
	return c.waitForImage(ctx, accepted.ID, options.Interval)
}
func (c *Client) WaitForImage(ctx context.Context, recordID string, options PollOptions) (ImageGenerationStatusRecord, error) {
	ctx, cancel, err := options.bounded(ctx)
	if err != nil {
		return ImageGenerationStatusRecord{}, err
	}
	defer cancel()
	return c.waitForImage(ctx, recordID, options.Interval)
}
func (c *Client) waitForImage(ctx context.Context, recordID string, interval time.Duration) (ImageGenerationStatusRecord, error) {
	for {
		record, err := c.GetImageGeneration(ctx, GetImageGenerationParams{Record: recordID})
		if err != nil {
			return record, err
		}
		if record.ID != recordID {
			return record, &ProtocolError{"image record identity mismatch"}
		}
		switch record.Status {
		case "succeeded", "failed", "cancelled":
			return record, ctx.Err()
		case "accepted", "queued", "processing", "submitted":
		default:
			return record, &ProtocolError{"unknown image lifecycle status"}
		}
		if err = wait(ctx, interval); err != nil {
			return record, err
		}
	}
}
