package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataWorkerPool() *schema.Resource {
	return &schema.Resource{
		Read: dataWorkerPoolRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "name of the worker pool",
				Computed:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "description of the worker pool",
				Computed:    true,
			},
			"worker_pool_id": &schema.Schema{
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
	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for worker pool")
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		return errors.New("worker pool not found")
	}

	d.SetId(workerPoolID)
	d.Set("name", workerPool.Name)
	d.Set("description", workerPool.Description)

	return nil
}
