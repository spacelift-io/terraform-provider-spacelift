package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataIdpGroupMapping() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_idp_group_mapping` represents a data source for retrieving " +
			"information about an existing IdP group mapping in Spacelift.",

		ReadContext: dataIdpGroupMappingRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Name of the IdP group as defined in the SSO provider",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the IdP group mapping",
			},
			"policy": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of access rules for the IdP group.",
				Deprecated:  "IdP group policies will be removed in a future version. Please use the `spacelift_role_attachment` resource to manage group roles instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID (slug) of the space the IdP group mapping has access to",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of access to the space. Possible values are: READ, WRITE, ADMIN",
						},
					},
				},
				Set: userPolicyHash,
			},
		},
	}
}

func dataIdpGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	name := d.Get("name").(string)

	var query struct {
		UserGroups []structs.UserGroup `graphql:"managedUserGroups"`
	}
	if err := meta.(*internal.Client).Query(ctx, "ManagedUserGroupsRead", &query, map[string]any{}); err != nil {
		return diag.Errorf("could not query for IdP group mappings: %v", err)
	}

	var userGroup *structs.UserGroup
	for i := range query.UserGroups {
		if query.UserGroups[i].Name == name {
			userGroup = &query.UserGroups[i]
			break
		}
	}

	if userGroup == nil {
		return diag.Errorf("could not find IdP group mapping with name %q", name)
	}

	d.SetId(userGroup.ID)
	d.Set("description", userGroup.Description)

	var accessList []any
	for _, a := range userGroup.AccessRules {
		accessList = append(accessList, map[string]any{
			"space_id": a.Space,
			"role":     a.SpaceAccessLevel,
		})
	}
	d.Set("policy", accessList)

	return nil
}
