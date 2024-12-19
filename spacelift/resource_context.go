package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceContext() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_context` represents a Spacelift **context** - " +
			"a collection of configuration elements (either environment variables or " +
			"mounted files) that can be administratively attached to multiple " +
			"stacks (`spacelift_stack`) or modules (`spacelift_module`) using " +
			"a context attachment (`spacelift_context_attachment`)`",

		CreateContext: resourceContextCreate,
		ReadContext:   resourceContextRead,
		UpdateContext: resourceContextUpdate,
		DeleteContext: resourceContextDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"after_apply": {
				Type:        schema.TypeList,
				Description: "List of after-apply scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_destroy": {
				Type:        schema.TypeList,
				Description: "List of after-destroy scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_init": {
				Type:        schema.TypeList,
				Description: "List of after-init scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_perform": {
				Type:        schema.TypeList,
				Description: "List of after-perform scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_plan": {
				Type:        schema.TypeList,
				Description: "List of after-plan scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"after_run": {
				Type:        schema.TypeList,
				Description: "List of after-run scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_destroy": {
				Type:        schema.TypeList,
				Description: "List of before-destroy scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_perform": {
				Type:        schema.TypeList,
				Description: "List of before-perform scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"before_plan": {
				Type:        schema.TypeList,
				Description: "List of before-plan scripts",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form context description for users",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "The labels of the context. To leverage the `autoattach` magic label, ensure your label follows the naming convention: `autoattach:<your-label-name>`",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the context - should be unique in one account",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the context is in",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceContextCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateContext structs.Context `graphql:"contextCreateV2(input: $input)"`
	}

	input := structs.ContextInput{
		Name:  toString(d.Get("name")),
		Hooks: buildHooksInput(d),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		input.Space = graphql.NewID(spaceID)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		input.Labels = &labels
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "ContextCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create context: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateContext.ID)

	return resourceContextRead(ctx, d, meta)
}

func resourceContextRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *structs.Context `graphql:"context(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "ContextRead", &query, variables); err != nil {
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

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range context.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)
	d.Set("space_id", context.Space)

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

func resourceContextUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateContext structs.Context `graphql:"contextUpdateV2(id: $id, input: $input)"`
	}

	input := structs.ContextInput{
		Name:  toString(d.Get("name")),
		Hooks: buildHooksInput(d),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		input.Space = graphql.NewID(spaceID)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		input.Labels = &labels
	}

	var ret diag.Diagnostics

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ContextUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update context: %v", internal.FromSpaceliftError(err))...)
	}

	ret = append(ret, resourceContextRead(ctx, d, meta)...)

	return ret
}

func resourceContextDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteContext *structs.Context `graphql:"contextDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "ContextDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete context: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func buildHooksInput(d *schema.ResourceData) *structs.HooksInput {
	return &structs.HooksInput{
		AfterApply:    gqlStringList(d, "after_apply"),
		AfterDestroy:  gqlStringList(d, "after_destroy"),
		AfterInit:     gqlStringList(d, "after_init"),
		AfterPerform:  gqlStringList(d, "after_perform"),
		AfterPlan:     gqlStringList(d, "after_plan"),
		AfterRun:      gqlStringList(d, "after_run"),
		BeforeApply:   gqlStringList(d, "before_apply"),
		BeforeDestroy: gqlStringList(d, "before_destroy"),
		BeforeInit:    gqlStringList(d, "before_init"),
		BeforePerform: gqlStringList(d, "before_perform"),
		BeforePlan:    gqlStringList(d, "before_plan"),
	}
}

func gqlStringList(d *schema.ResourceData, key string) []graphql.String {
	ret := []graphql.String{}

	if list, ok := d.GetOk(key); ok {
		for _, item := range list.([]interface{}) {
			ret = append(ret, graphql.String(item.(string)))
		}
	}

	return ret
}
