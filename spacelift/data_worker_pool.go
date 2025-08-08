package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataWorkerPool() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_worker_pool` represents a worker pool assigned to the " +
			"Spacelift account.",

		ReadContext: dataWorkerPoolRead,

		Schema: map[string]*schema.Schema{
			"config": {
				Type:        schema.TypeString,
				Description: "credentials necessary to connect WorkerPool's workers to the control plane",
				Computed:    true,
				Sensitive:   true,
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
			"worker_pool_id": {
				Type:             schema.TypeString,
				Description:      "ID of the worker pool",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the worker pool is in",
				Computed:    true,
			},
			"drift_detection_run_limit": {
				Type:        schema.TypeInt,
				Description: "Limit of how many concurrent drift detection runs are allowed per worker pool",
				Computed:    true,
			},
		},
	}
}

func dataWorkerPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPool(id: $id)"`
	}

	workerPoolID := d.Get("worker_pool_id").(string)

	variables := map[string]interface{}{"id": toID(workerPoolID)}
	if err := meta.(*internal.Client).Query(ctx, "WorkerPoolRead", &query, variables); err != nil {
		return diag.Errorf("could not query for worker pool: %v", err)
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		return diag.Errorf("worker pool not found")
	}

	d.SetId(workerPoolID)
	d.Set("name", workerPool.Name)
	d.Set("config", workerPool.Config)
	d.Set("space_id", workerPool.Space)

	if workerPool.Description != nil {
		d.Set("description", *workerPool.Description)
	} else {
		d.Set("description", nil)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range workerPool.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if workerPool.DriftDetectionRunLimit != nil {
		d.Set("drift_detection_run_limit", *workerPool.DriftDetectionRunLimit)
	}

	return nil
}
