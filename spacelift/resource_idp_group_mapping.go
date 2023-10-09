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

func resourceIdpGroupMapping() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_idp_group_mapping` represents a mapping (binding) between a user group (as provided by IdP) " +
			"and a Spacelift User Management Policy. If you assign permissions (a Policy) to a user group, all users in the group " +
			"will have those permissions unless the user's permissions are higher than the group's permissions.",
		CreateContext: resourceIdpGroupMappingCreate,
		ReadContext:   resourceIdpGroupMappingRead,
		UpdateContext: resourceIdpGroupMappingUpdate,
		DeleteContext: resourceIdpGroupMappingDelete,

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
			"policy": {
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
						"role": {
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

func resourceIdpGroupMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// send a create query to the API
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
		return diag.Errorf("could not create user group mapping %v: %v", toString(d.Get("name")), internal.FromSpaceliftError(err))
	}

	// set the ID in TF state
	d.SetId(mutation.UserGroup.ID)

	// fetch from remote and write to TF state
	return resourceIdpGroupMappingRead(ctx, d, meta)
}

func resourceIdpGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// send a read query to the API
	var query struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroup(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "ManagedUserGroupRead", &query, variables); err != nil {
		return diag.Errorf("could not query for user group mapping: %v", err)
	}

	// if the mapping is not found on the Spacelift side, delete it from the TF state
	userGroup := query.UserGroup
	if userGroup == nil {
		d.SetId("")
		return nil
	}

	// if found, update the TF state
	d.Set("name", userGroup.Name)
	var accessList []interface{}
	for _, a := range userGroup.AccessRules {
		accessList = append(accessList, map[string]interface{}{
			"space_id": a.Space,
			"role":     a.SpaceAccessLevel,
		})
	}
	d.Set("policy", accessList)

	return nil

}

func resourceIdpGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ret diag.Diagnostics

	// send an update query to the API
	var mutation struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroupUpdate(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": structs.ManagedUserGroupUpdateInput{
			ID:          toID(d.Id()),
			AccessRules: getAccessRules(d),
		},
	}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update user group mapping: %v", internal.FromSpaceliftError(err))...)
	}

	// send a read query to the API
	ret = append(ret, resourceIdpGroupMappingRead(ctx, d, meta)...)

	return ret
}

func resourceIdpGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// send a delete query to the API
	var mutation struct {
		UserGroup *structs.UserGroup `graphql:"managedUserGroupDelete(id: $id)"`
	}
	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete user group mapping: %v", internal.FromSpaceliftError(err))
	}

	// if the mapping was successfully removed from the Spacelift side, delete it from the TF state
	d.SetId("")

	return nil
}

func getAccessRules(d *schema.ResourceData) []structs.SpaceAccessRuleInput {
	var accessRules []structs.SpaceAccessRuleInput
	for _, a := range d.Get("policy").([]interface{}) {
		access := a.(map[string]interface{})
		accessRules = append(accessRules, structs.SpaceAccessRuleInput{
			Space:            toID(access["space_id"]),
			SpaceAccessLevel: structs.SpaceAccessLevel(access["role"].(string)),
		})
	}
	return accessRules
}
