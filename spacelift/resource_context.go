package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceContext() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContextCreate,
		ReadContext:   resourceContextRead,
		UpdateContext: resourceContextUpdate,
		DeleteContext: resourceContextDelete,

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

func resourceContextCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not create context: %v", err)
	}

	d.SetId(mutation.CreateContext.ID)

	return resourceContextRead(ctx, d, meta)
}

func resourceContextRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, &query, variables); err != nil {
		return diag.Errorf("could not query for context: %v", err)
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

func resourceContextUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update context: %v", err)...)
	}

	ret = append(ret, resourceContextRead(ctx, d, meta)...)

	return ret
}

func resourceContextDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteContext *structs.Context `graphql:"contextDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not delete context: %v", err)
	}

	d.SetId("")

	return nil
}
