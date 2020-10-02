package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataWorkerPool() *schema.Resource {
	return &schema.Resource{
		Read: dataWorkerPoolRead,
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
				Type:        schema.TypeString,
				Description: "ID of the worker pool",
				Required:    true,
			},
		},
	}
}

func dataWorkerPoolRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPool(id: $id)"`
	}

	workerPoolID := d.Get("worker_pool_id").(string)

	variables := map[string]interface{}{"id": toID(workerPoolID)}
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for worker pool")
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		return errors.New("worker pool not found")
	}

	d.SetId(workerPoolID)
	d.Set("name", workerPool.Name)
	d.Set("config", workerPool.Config)

	if workerPool.Description != nil {
		d.Set("description", *workerPool.Description)
	} else {
		d.Set("description", nil)
	}

	return nil
}
