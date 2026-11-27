package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewContextListResource() list.ListResource {
	return &sdkListResource[structs.Context]{
		typeName: "spacelift_context", singular: "context", plural: "contexts",
		resource: resourceContext,
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.Context, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.Context] `graphql:"searchContexts(input: $input)"`
			}
			err := client.Query(ctx, "ContextList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		keep:     func(_ *listFilter, context *structs.Context) bool { return hasULID(context.ULID) },
		name:     func(context *structs.Context) string { return context.Name },
		populate: populateContext,
	}
}
