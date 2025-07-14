package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataRole() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_role` represents a Spacelift **role** - " +
			"a collection of permissions that can be assigned to IdP groups or API keys " +
			"to control access to Spacelift resources and operations.\n\n" +
			"You can either filter roles by their unique identifier (`role_id`) " +
			"or by their human-readable name (`name`).\n\n" +
			"**Note:** you must have admin access to the `root` Space in order to retrieve roles.",

		ReadContext: dataRoleRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:          schema.TypeString,
				Description:   "Unique identifier (ULID) of the role. Can be used to filter roles.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Description:   "Human-readable, free-form name of the role. Can be used to filter roles.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"role_id"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Human-readable, free-form description of the role",
				Computed:    true,
			},
			"actions": {
				Type:        schema.TypeSet,
				Description: "List of actions (permissions) associated with the role",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"is_system": {
				Type:        schema.TypeBool,
				Description: "Whether the role is a system role (Space admin, Space writer, Space reader). The 3 system roles are created by default and cannot be deleted or modified.",
				Computed:    true,
			},
		},
	}
}

func dataRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	filterToRoleID := d.Get("role_id").(string)
	filterToRoleName := d.Get("name").(string)
	if filterToRoleID == "" && filterToRoleName == "" {
		return diag.Errorf("either 'role_id' or 'name' must be specified to read a role")
	}

	var query struct {
		Roles []*structs.Role `graphql:"roles"`
	}

	if err := meta.(*internal.Client).Query(ctx, "ReadAllRoles", &query, nil); err != nil {
		return diag.Errorf("could not query for role: %v", err)
	}

	allRoles := query.Roles
	if len(allRoles) == 0 {
		return diag.Errorf("no roles found. Ensure you have root admin access to the Spacelift API.")
	}

	var role *structs.Role

	for _, r := range allRoles {
		if r.ID == filterToRoleID {
			role = r
			break
		}
		if r.Name == filterToRoleName {
			role = r
			break
		}
	}

	if role == nil {
		if filterToRoleID != "" {
			return diag.Errorf("role with ID %s not found", filterToRoleID)
		}

		return diag.Errorf("role with name %s not found", filterToRoleName)
	}

	d.SetId(role.ID)
	d.Set("role_id", role.ID)
	d.Set("name", role.Name)
	d.Set("description", role.Description)
	d.Set("is_system", role.IsSystem)

	actions := schema.NewSet(schema.HashString, []interface{}{})
	for _, action := range role.Actions {
		actions.Add(string(action))
	}
	d.Set("actions", actions)

	return nil
}
