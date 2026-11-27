package spacelift

import (
	"context"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewSpaceListResource() list.ListResource {
	return &sdkListResource[structs.Space]{
		typeName: "spacelift_space", singular: "space", plural: "spaces",
		resource:       resourceSpace,
		spaceAttribute: "parent_space_id",
		attributes: func() map[string]listschema.Attribute {
			attributes := listFilterAttributes("spaces")
			delete(attributes, "space_id")
			attributes["parent_space_id"] = listschema.StringAttribute{
				Description: "Only list the child spaces of this space.",
				Optional:    true,
			}
			attributes["include_child_spaces"] = listschema.BoolAttribute{
				Description: "With `parent_space_id`, also list the children's own child spaces, at any depth.",
				Optional:    true,
			}
			return attributes
		},
		// The space search matches a space's own path, so it searches the parent's
		// whole subtree and keep checks the parent.
		predicates: func(_ context.Context, _ tfsdk.Config, filter *listFilter) ([]search.SearchQueryPredicate, fwdiag.Diagnostics) {
			subtree := *filter
			subtree.includeChildSpaces = true
			return subtree.predicates(), nil
		},
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.Space, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.Space] `graphql:"searchSpaces(input: $input)"`
			}
			err := client.Query(ctx, "SpaceList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		keep: func(filter *listFilter, space *structs.Space) bool {
			// root and legacy are built in, never managed by Terraform.
			if space.ID == "root" || space.ID == "legacy" || space.ParentSpace == nil {
				return false
			}
			return filter.matches(space.Name, *space.ParentSpace, space.Labels)
		},
		name:     func(space *structs.Space) string { return space.Name },
		populate: populateSpace,
	}
}
