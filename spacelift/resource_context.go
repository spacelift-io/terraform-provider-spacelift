package spacelift

import (
	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceContextCreate,
		Read:   resourceContextRead,
		Update: resourceContextUpdate,
		Delete: resourceContextDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form context description for users",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the context - should be unique in one account",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceContextCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreateContext structs.Context `graphql:"contextCreate(name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"name":        toString(d.Get("name")),
		"description": (*graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create context")
	}

	d.SetId(mutation.CreateContext.ID)

	return resourceContextRead(d, meta)
}

func resourceContextRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for context")
	}

	context := query.Context
	if context == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", context.Name)

	if description := context.Description; description != nil {
		d.Set("description", *description)
	}

	return nil
}

func resourceContextUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		UpdateContext structs.Context `graphql:"contextUpdate(id: $id, name: $name, description: $description)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Id()),
		"name":        toString(d.Get("name")),
		"description": (*graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(meta.(*internal.Client).Mutate(&mutation, variables), "could not update context"))
	acc.Push(errors.Wrap(resourceContextRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourceContextDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteContext *structs.Context `graphql:"contextDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete context")
	}

	d.SetId("")

	return nil
}
