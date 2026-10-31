package envai

import (
	"context"
	"iter"
)

// CompletionPages follows numeric Laravel pages on the configured API origin.
// It never follows a server-supplied URL (and cannot leak credentials to one).
// Stop iteration to stop fetching; errors are yielded once and end iteration.
func (c *Client) CompletionPages(ctx context.Context, request ListCompletionsParams) iter.Seq2[CompletionCollection, error] {
	return func(yield func(CompletionCollection, error) bool) {
		page := int64(1)
		if request.Page.Set {
			page = request.Page.Value
		}
		if page < 1 {
			yield(CompletionCollection{}, &ProtocolError{"page must be positive"})
			return
		}
		for {
			request.Page = Present(page)
			result, err := c.ListCompletions(ctx, request)
			if err != nil {
				yield(result, err)
				return
			}
			if !result.Meta.CurrentPage.Set || result.Meta.CurrentPage.Null || result.Meta.CurrentPage.Value != page || !result.Meta.LastPage.Set || result.Meta.LastPage.Null || result.Meta.LastPage.Value < page {
				yield(result, &ProtocolError{"invalid pagination metadata"})
				return
			}
			if !yield(result, nil) || page == result.Meta.LastPage.Value {
				return
			}
			page++
		}
	}
}
