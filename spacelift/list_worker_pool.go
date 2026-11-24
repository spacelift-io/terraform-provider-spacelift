package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewWorkerPoolListResource() list.ListResource {
	return &sdkListResource[structs.WorkerPool]{
		typeName: "spacelift_worker_pool", singular: "worker pool", plural: "worker pools",
		resource: resourceWorkerPool,
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.WorkerPool, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.WorkerPool] `graphql:"searchWorkerPools(input: $input)"`
			}
			err := client.Query(ctx, "WorkerPoolList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		name:     func(pool *structs.WorkerPool) string { return pool.Name },
		populate: populateWorkerPool,
	}
}
