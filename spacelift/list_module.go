package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewModuleListResource() list.ListResource {
	return &sdkListResource[structs.Module]{
		typeName: "spacelift_module", singular: "module", plural: "modules",
		resource: resourceModule,
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.Module, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.Module] `graphql:"searchModules(input: $input)"`
			}
			err := client.Query(ctx, "ModuleList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		keep:     func(_ *listFilter, module *structs.Module) bool { return hasULID(module.ULID) },
		name:     func(module *structs.Module) string { return module.Name },
		populate: populateModule,
	}
}
