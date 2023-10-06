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

var validAccessLevels = []string{
	"READ",
	"WRITE",
	"ADMIN",
}

func resourceUserGroup() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_user_group` represents a Spacelift **user group** - " +
			"a collection of users as provided by your Identity Provider (IdP). " +
			"If you assign permissions to a user group, all users in the group " +
			"will have those permissions unless the user's permissions are higher than " +
			"the group's permissions.",
		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the user group - should be unique in one account",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"access": {
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:             schema.TypeString,
							Description:      "ID (slug) of the space the user group has access to",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"level": {
							Type: schema.TypeString,
							Description: "Type of access to the space. Possible values are: " +
								"READ, WRITE, ADMIN",
							Required:     true,
							ValidateFunc: validation.StringInSlice(validAccessLevels, false),
						},
					},
				},
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroupCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.ManagedUserGroupCreateInput{
			Name:        toString(d.Get("name")),
			AccessRules: getAccessRules(d),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create user group %v: %v", toString(d.Get("name")), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.UserGroup.ID)

	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroup(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "ManagedUserGroupRead", &query, variables); err != nil {
		return diag.Errorf("could not query for user group: %v", err)
	}

	userGroup := query.UserGroup
	if userGroup == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", userGroup.Name)

	var accessList []interface{}

	for _, a := range userGroup.AccessRules {
		accessList = append(accessList, map[string]interface{}{
			"space_id": a.Space,
			"level":    a.SpaceAccessLevel,
		})
	}

	d.Set("access", accessList)

	return nil

}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroupUpdate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.ManagedUserGroupUpdateInput{
			ID:          toID(d.Id()),
			AccessRules: getAccessRules(d),
		},
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update user group: %v", internal.FromSpaceliftError(err))...)
	}

	ret = append(ret, resourceUserGroupRead(ctx, d, meta)...)

	return ret
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroupDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete user group: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func getAccessRules(d *schema.ResourceData) []structs.SpaceAccessRuleInput {
	var accessRules []structs.SpaceAccessRuleInput
	for _, a := range d.Get("access").([]interface{}) {
		access := a.(map[string]interface{})
		accessRules = append(accessRules, structs.SpaceAccessRuleInput{
			Space:            toID(access["space_id"]),
			SpaceAccessLevel: structs.SpaceAccessLevel(access["level"].(string)),
		})
	}
	return accessRules
}
