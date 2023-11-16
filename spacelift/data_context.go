package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataContext() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_context` represents a Spacelift **context** - " +
			"a collection of configuration elements (either environment variables or " +
			"mounted files) that can be administratively attached to multiple " +
			"stacks (`spacelift_stack`) or modules (`spacelift_module`) using " +
			"a context attachment (`spacelift_context_attachment`)`",

		ReadContext: dataContextRead,

		Schema: map[string]*schema.Schema{
			"after_apply": {
				Type:        schema.TypeList,
				Description: "List of after-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_destroy": {
				Type:        schema.TypeList,
				Description: "List of after-destroy scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_init": {
				Type:        schema.TypeList,
				Description: "List of after-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_perform": {
				Type:        schema.TypeList,
				Description: "List of after-perform scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_plan": {
				Type:        schema.TypeList,
				Description: "List of after-plan scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"after_run": {
				Type:        schema.TypeList,
				Description: "List of after-run scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_destroy": {
				Type:        schema.TypeList,
				Description: "List of before-destroy scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_perform": {
				Type:        schema.TypeList,
				Description: "List of before-perform scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"before_plan": {
				Type:        schema.TypeList,
				Description: "List of before-plan scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
			},
			"context_id": {
				Type:             schema.TypeString,
				Description:      "immutable ID (slug) of the context",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form context description for users",
				Computed:    true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the context",
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the context is in",
				Computed:    true,
			},
		},
	}
}

func dataContextRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("context_id"))}
	if err := meta.(*internal.Client).Query(ctx, "ContextRead", &query, variables); err != nil {
		return diag.Errorf("could not query for context: %v", err)
	}

	context := query.Context
	if context == nil {
		return diag.Errorf("context not found")
	}

	d.SetId(context.ID)
	d.Set("name", context.Name)
	d.Set("space_id", context.Space)

	if context.Description != nil {
		d.Set("description", *context.Description)
	} else {
		d.Set("description", nil)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range context.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	d.Set("after_apply", context.Hooks.AfterApply)
	d.Set("after_destroy", context.Hooks.AfterDestroy)
	d.Set("after_init", context.Hooks.AfterInit)
	d.Set("after_perform", context.Hooks.AfterPerform)
	d.Set("after_plan", context.Hooks.AfterPlan)
	d.Set("after_run", context.Hooks.AfterRun)
	d.Set("before_apply", context.Hooks.BeforeApply)
	d.Set("before_destroy", context.Hooks.BeforeDestroy)
	d.Set("before_init", context.Hooks.BeforeInit)
	d.Set("before_perform", context.Hooks.BeforePerform)
	d.Set("before_plan", context.Hooks.BeforePlan)

	return nil
}
