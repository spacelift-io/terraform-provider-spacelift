package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_user` represents a Spacelift user. ",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Description: "Username of the user",
				Required:    true,
			},
			"policy": {
				Type: schema.TypeList,
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	// send a create query to the API
	var mutation struct {
		User *structs.User `graphql:`
	}

}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	// send a read query to the API
	var query struct {
		User *structs.User `graphql:"managedUser(id: $id)"`
	}
	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := i.(*internal.Client).Query(ctx, "ManagedUserRead", &query, variables); err != nil {
		return diag.Errorf("could not query for user: %v", err)
	}

	// if the user was not found on the Spacelift side, delete it from TF state
	if query.User == nil {
		d.SetId("")
		return nil
	}

	// if found, update the TF state
	d.Set("username", query.User.Username)
	var accessList []interface{}
	for _, a := range query.User.AccessRules {
		accessList = append(accessList, map[string]interface{}{
			"space_id": a.Space,
			"role":     a.SpaceAccessLevel,
		})
	}
	d.Set("policy", accessList)

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	var ret diag.Diagnostics

	// send an update query to the API
	var mutation struct {
		User *structs.User `graphql:"managedUserUpdate(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": structs.ManagedUserUpdateInput{
			ID:          toID(d.Id()),
			AccessRules: getAccessRules(d),
		},
	}
	if err := i.(*internal.Client).Mutate(ctx, "ManagedUserUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update user %s: %v", d.Id(), internal.FromSpaceliftError(err))...)
	}

	// send a read query to the API
	ret = append(ret, resourceUserRead(ctx, d, i)...)

	return ret

}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	// send a delete query to the API
	var mutation struct {
		User *structs.User `graphql:"managedUserDelete(id: $id)"`
	}
	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := i.(*internal.Client).Mutate(ctx, "ManagedUserDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete user %s: %v", d.Id(), internal.FromSpaceliftError(err))
	}

	// if the user was deleted, remove it from the TF state as well
	d.SetId("")

	return nil
}
