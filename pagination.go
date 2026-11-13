package envai

import (
	"context"
	"iter"
)

func numericPages[T any](ctx context.Context, start Optional[int64], advance func(int64), fetch func(context.Context) (T, error), progress func(T) (Optional[int64], Optional[int64])) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		page := int64(1)
		if start.Set {
			page = start.Value
		}
		if page < 1 {
			yield(zero, &ProtocolError{"page must be positive"})
			return
		}
		for {
			if err := ctx.Err(); err != nil {
				yield(zero, err)
				return
			}
			advance(page)
			result, err := fetch(ctx)
			if err != nil {
				yield(result, err)
				return
			}
			current, last := progress(result)
			if !current.Set || current.Null || current.Value != page || !last.Set || last.Null || last.Value < page {
				yield(result, &ProtocolError{"invalid pagination metadata"})
				return
			}
			if !yield(result, nil) || page == last.Value {
				return
			}
			page++
		}
	}
}

// CompletionPages follows numeric Laravel pages on the configured API origin.
// It never follows a server-supplied URL (and cannot leak credentials to one).
// Stop iteration to stop fetching; errors are yielded once and end iteration.
func (c *Client) CompletionPages(ctx context.Context, request ListCompletionsParams) iter.Seq2[CompletionCollection, error] {
	return numericPages(ctx, request.Page,
		func(page int64) { request.Page = Present(page) },
		func(ctx context.Context) (CompletionCollection, error) { return c.ListCompletions(ctx, request) },
		func(result CompletionCollection) (Optional[int64], Optional[int64]) {
			return result.Meta.CurrentPage, result.Meta.LastPage
		})
}

func (c *Client) ProjectUsagePages(ctx context.Context, request GetProjectUsageParams) iter.Seq2[UsageReportCollection, error] {
	return numericPages(ctx, request.Page,
		func(page int64) { request.Page = Present(page) },
		func(ctx context.Context) (UsageReportCollection, error) { return c.GetProjectUsage(ctx, request) },
		func(result UsageReportCollection) (Optional[int64], Optional[int64]) {
			return result.Meta.CurrentPage, result.Meta.LastPage
		})
}

func (c *Client) ProjectLogPages(ctx context.Context, request GetProjectLogsParams) iter.Seq2[ProjectLogCollection, error] {
	return numericPages(ctx, request.Page,
		func(page int64) { request.Page = Present(page) },
		func(ctx context.Context) (ProjectLogCollection, error) { return c.GetProjectLogsEvents(ctx, request) },
		func(result ProjectLogCollection) (Optional[int64], Optional[int64]) {
			return result.Meta.CurrentPage, result.Meta.LastPage
		})
}
