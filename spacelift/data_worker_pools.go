package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataWorkerPools() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataWorkerPoolsRead,
		Schema: map[string]*schema.Schema{
			"worker_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"worker_pool_id": {
							Type:        schema.TypeString,
							Description: "ID of the worker pool",
							Computed:    true,
						},
						"config": {
							Type:        schema.TypeString,
							Description: "credentials necessary to connect WorkerPool's workers to the control plane",
							Sensitive:   true,
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "name of the worker pool",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "description of the worker pool",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataWorkerPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var query struct {
		WorkerPools []*structs.WorkerPool `graphql:"workerPools()"`
	}
	variables := map[string]interface{}{}

	if err := meta.(*internal.Client).Query(ctx, "WorkerPoolsRead", &query, variables); err != nil {
		return diag.Errorf("could not query for worker pools: %v", err)
	}

	d.SetId("spacelift-worker-pools")

	workerPools := query.WorkerPools
	if workerPools == nil {
		d.Set("worker_pools", nil)
		return nil
	}

	wps := flattenDataWorkerPoolsList(workerPools)
	if err := d.Set("worker_pools", wps); err != nil {
		d.SetId("")
		return diag.Errorf("could not set worker pools: %v", err)
	}

	return nil
}

func flattenDataWorkerPoolsList(workerPools []*structs.WorkerPool) []map[string]interface{} {
	wps := make([]map[string]interface{}, len(workerPools))

	for index, wp := range workerPools {
		var description *string

		if wp.Description != nil {
			description = wp.Description
		} else {
			description = nil
		}

		wps[index] = map[string]interface{}{
			"worker_pool_id": wp.ID,
			"name":           wp.Name,
			"config":         wp.Config,
			"description":    description,
		}
	}

	return wps
}
