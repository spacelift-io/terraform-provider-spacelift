package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewStackListResource() list.ListResource {
	return &sdkListResource[structs.Stack]{
		typeName: "spacelift_stack", singular: "stack", plural: "stacks",
		resource: resourceStack,
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.Stack, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.Stack] `graphql:"searchStacks(input: $input)"`
			}
			err := client.Query(ctx, "StackList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		keep:     func(_ *listFilter, stack *structs.Stack) bool { return hasULID(stack.ULID) },
		name:     func(stack *structs.Stack) string { return stack.Name },
		populate: structs.PopulateStack,
	}
}
