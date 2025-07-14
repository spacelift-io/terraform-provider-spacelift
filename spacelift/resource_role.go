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

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_role` represents a Spacelift **role** - " +
			"a collection of permissions that can be assigned to IdP groups or API keys " +
			"to control access to Spacelift resources and operations.\n\n" +
			"**Note:** you must have admin access to the `root` Space in order to create or mutate roles.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Unique identifier (ULID) of the role",
				Computed:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Human-readable, free-form name of the role",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Human-readable, free-form description of the role",
				Optional:    true,
			},
			"actions": {
				Type:        schema.TypeSet,
				Description: "List of actions (permissions) associated with the role.",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.ValidateAction,
				},
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateRole structs.Role `graphql:"roleCreate(input: $input)"`
	}

	input := structs.RoleInput{
		Name:    toString(d.Get("name")),
		Actions: actionsToGraphQLStringList(d.Get("actions").(*schema.Set)),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "RoleCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create role: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateRole.ID)

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Role *structs.Role `graphql:"role(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "RoleRead", &query, variables); err != nil {
		return diag.Errorf("could not query for role: %v", err)
	}

	role := query.Role
	if role == nil {
		d.SetId("")
		return nil
	}

	d.Set("id", role.ID)
	d.Set("name", role.Name)
	d.Set("description", role.Description)

	actions := schema.NewSet(schema.HashString, []interface{}{})
	for _, action := range role.Actions {
		actions.Add(string(action))
	}
	d.Set("actions", actions)

	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateRole structs.Role `graphql:"roleUpdate(id: $id, input: $input)"`
	}

	input := structs.RoleUpdateInput{}

	if d.HasChange("name") {
		input.Name = toOptionalString(d.Get("name"))
	}

	if d.HasChange("description") {
		input.Description = toOptionalString(d.Get("description"))
	}

	if d.HasChange("actions") {
		actions := actionsToGraphQLStringList(d.Get("actions").(*schema.Set))
		input.Actions = &actions
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "RoleUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update role: %v", internal.FromSpaceliftError(err))
	}

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteRole *structs.Role `graphql:"roleDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "RoleDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete role: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func actionsToGraphQLStringList(actionSet *schema.Set) []graphql.String {
	ret := []graphql.String{}

	for _, action := range actionSet.List() {
		ret = append(ret, graphql.String(action.(string)))
	}

	return ret
}
