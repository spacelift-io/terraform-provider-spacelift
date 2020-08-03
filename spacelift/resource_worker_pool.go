package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceWorkerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkerPoolCreate,
		Read:   resourceWorkerPoolRead,
		Update: resourceWorkerPoolUpdate,
		Delete: resourceWorkerPoolDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"config": &schema.Schema{
				Type:        schema.TypeString,
				Description: "credentials necessary to connect WorkerPool's workers to the control plane",
				Computed:    true,
				Sensitive:   true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "name of the worker pool",
				Required:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "description of the worker pool",
				Optional:    true,
			},
		},
	}
}

func resourceWorkerPoolCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	var mutation struct {
		WorkerPoolConfig struct {
			Config     string             `graphql:"config"`
			WorkerPool structs.WorkerPool `graphql:"workerPool"`
		} `graphql:"workerPoolCreate(name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create worker pool")
	}

	d.SetId(mutation.WorkerPoolConfig.WorkerPool.ID)
	d.Set("config", mutation.WorkerPoolConfig.Config)
	d.Set("name", mutation.WorkerPoolConfig.WorkerPool.Name)

	if description := mutation.WorkerPoolConfig.WorkerPool.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceWorkerPoolRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPool(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for worker pool")
	}

	workerPool := query.WorkerPool
	if workerPool == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", workerPool.Name)
	if description := workerPool.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceWorkerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	var mutation struct {
		WorkerPool structs.WorkerPool `graphql:"workerPoolUpdate(id: $id, name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Id()),
		"name":        graphql.String(name),
		"description": (*graphql.String)(nil),
	}

	if desc, ok := d.GetOk("description"); ok {
		variables["description"] = graphql.String(desc.(string))
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update worker pool")
	}

	return nil
}

func resourceWorkerPoolDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		WorkerPool *structs.WorkerPool `graphql:"workerPoolDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete worker pool")
	}

	d.SetId("")

	return nil
}
