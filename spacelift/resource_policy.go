package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

var policyTypes = []string{
	"ACCESS",
	"APPROVAL",
	"GIT_PUSH",
	"INITIALIZATION",
	"LOGIN",
	"PLAN",
	"STACK_ACCESS", // deprecated
	"TASK",
	"TASK_RUN",       // deprecated
	"TERRAFORM_PLAN", // deprecated
	"TRIGGER",
	"NOTIFICATION",
}

var policyEngineTypes = []string{
	"REGO_V0",
	"REGO_V1",
}

// This is a map of new policy type names to the ones they are replacing.
var typeNameReplacements = map[string]string{
	"ACCESS": "STACK_ACCESS",
	"PLAN":   "TERRAFORM_PLAN",
	"TASK":   "TASK_RUN",
}

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_policy` represents a Spacelift **policy** - a collection of " +
			"customer-defined rules that are applied by Spacelift at one of the " +
			"decision points within the application.",

		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the policy - should be unique in one account",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"body": {
				Type:             schema.TypeString,
				Description:      "Body of the policy",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"labels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the policy is in",
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the policy. Possible values are `ACCESS`, `APPROVAL`, `GIT_PUSH`, `INITIALIZATION`, `LOGIN`, `PLAN`, `TASK`, `TRIGGER` and `NOTIFICATION`. Deprecated values are `STACK_ACCESS` (use `ACCESS` instead), `TASK_RUN` (use `TASK` instead), and `TERRAFORM_PLAN` (use `PLAN` instead).",
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// If the backend responds with a new name, but we still have the old
					// name defined or stored in the state, let's not do the replacement.
					if previous, ok := typeNameReplacements[new]; ok && previous == old {
						return true
					}
					next, ok := typeNameReplacements[old]
					return ok && next == new
				},
				ValidateFunc: validation.StringInSlice(
					policyTypes,
					false, // case-sensitive match
				),
			},
			"description": {
				Type:             schema.TypeString,
				Description:      "Description of the policy",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"engine_type": {
				Type:        schema.TypeString,
				Description: "Type of engine used to evaluate the policy. Possible values are `REGO_V0` and `REGO_V1`.",
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					policyEngineTypes,
					false,
				),
			},
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreatePolicy structs.Policy `graphql:"policyCreatev2(input: $input)"`
	}

	input := structs.NewPolicyCreateInput(toString(d.Get("name")), toString(d.Get("body")), structs.PolicyType(d.Get("type").(string)))

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		input.Labels = &labels
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		input.Space = graphql.NewID(spaceID)
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	engineType := d.Get("engine_type")
	if typeStr, ok := engineType.(string); ok && typeStr != "" {
		et := structs.PolicyEngineType(typeStr)
		input.EngineType = &et
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "PolicyCreateV2", &mutation, variables); err != nil {
		return diag.Errorf("could not create policy %v: %v", toString(d.Get("name")), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreatePolicy.ID)

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Policy *structs.Policy `graphql:"policy(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "PolicyRead", &query, variables); err != nil {
		return diag.Errorf("could not query for policy: %v", err)
	}

	policy := query.Policy
	if policy == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", policy.Name)
	d.Set("body", policy.Body)
	d.Set("type", policy.Type)
	d.Set("space_id", policy.Space)
	d.Set("description", policy.Description)
	d.Set("engine_type", policy.EngineType)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range policy.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if policy.Type == "TASK" || policy.Type == "INITIALIZATION" {
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "Policy type is deprecated, please use APPROVAL policy instead",
		}}
	}

	return nil
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdatePolicy structs.Policy `graphql:"policyUpdatev2(id: $id, input: $input)"`
	}

	input := structs.NewPolicyUpdateInput(toString(d.Get("name")), toString(d.Get("body")))

	if desc, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(desc)
	}

	engineType := d.Get("engine_type")
	if typeStr, ok := engineType.(string); ok && typeStr != "" {
		et := structs.PolicyEngineType(typeStr)
		input.EngineType = &et
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		input.Labels = &labels
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		input.Space = graphql.NewID(spaceID)
	}

	var ret diag.Diagnostics
	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "PolicyUpdateV2", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update policy: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeletePolicy *structs.Policy `graphql:"policyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "PolicyDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete policy: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
