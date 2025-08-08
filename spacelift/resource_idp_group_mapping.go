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
			"`spacelift_idp_group_mapping` represents a mapping between a group in an IdP " +
			"and Spacelift.\n" +
			"Note: The `policy` attribute, previously used to assign roles to the group, is deprecated. Use the `spacelift_role_attachment` resource to manage group roles instead.",
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
				Description:      "Name of the IdP group as defined in the SSO provider - should be unique per account",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"policy": {
				Type:        schema.TypeSet,
				Description: "List of access rules for the IdP group.",
				Optional:    true,
				Deprecated:  "IdP group policies will be removed in a future version. Please use the `spacelift_role_attachment` resource to manage group roles instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:             schema.TypeString,
							Description:      "ID (slug) of the space the IdP group mapping has access to",
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
				Set: userPolicyHash,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the IdP group mapping",
				Optional:    true,
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
			Description: toString(d.Get("description")),
			AccessRules: getAccessRules(d),
		},
	}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create IdP group mapping %v: %v", toString(d.Get("name")), internal.FromSpaceliftError(err))
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
		return diag.Errorf("could not query for IdP group mapping: %v", err)
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
	d.Set("description", userGroup.Description)
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
			Description: toString(d.Get("description")),
		},
	}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update IdP group mapping: %v", internal.FromSpaceliftError(err))...)
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
		return diag.Errorf("could not delete IdP group mapping: %v", internal.FromSpaceliftError(err))
	}

	// if the mapping was successfully removed from the Spacelift side, delete it from the TF state
	d.SetId("")

	return nil
}

func getAccessRules(d *schema.ResourceData) []structs.SpaceAccessRuleInput {
	accessRules := make([]structs.SpaceAccessRuleInput, 0)

	if policies, ok := d.Get("policy").(*schema.Set); ok {
		for _, a := range policies.List() {
			access := a.(map[string]interface{})
			accessRules = append(accessRules, structs.SpaceAccessRuleInput{
				Space:            toID(access["space_id"]),
				SpaceAccessLevel: structs.SpaceAccessLevel(access["role"].(string)),
			})
		}
	}

	return accessRules
}
