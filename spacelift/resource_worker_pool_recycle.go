package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func resourceWorkerPoolRecycle() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_worker_pool_recycle` represents a worker pool recycle operation in Spacelift. This resource triggers a recycle of all workers in the specified worker pool, causing them to be replaced with fresh instances.",

		CreateContext: resourceWorkerPoolRecycleCreate,
		ReadContext:   schema.NoopContext,
		DeleteContext: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to recycle",
				Required:    true,
				ForceNew:    true,
			},
			"keepers": {
				Description: "" +
					"Arbitrary map of values that, when changed, will trigger " +
					"recreation of the resource and thus a new worker pool recycle operation. " +
					"This allows you to control when the worker pool recycle should happen.",
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkerPoolRecycleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	workerPoolID := d.Get("worker_pool_id").(string)

	var mutation struct {
		WorkerPoolCycle graphql.Boolean `graphql:"workerPoolCycle(id: $workerPoolId)"`
	}

	variables := map[string]interface{}{
		"workerPoolId": graphql.ID(workerPoolID),
	}

	client := meta.(*internal.Client)
	if err := client.Mutate(ctx, "CycleWorkerPool", &mutation, variables); err != nil {
		return diag.Errorf("could not recycle worker pool: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("worker_pool_recycle_%s", workerPoolID))

	return nil
}
