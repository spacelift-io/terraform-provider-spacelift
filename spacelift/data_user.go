package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataUser() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_user` represents a data source for retrieving information " +
			"about an existing Spacelift user. This allows you to reference " +
			"user information in your Terraform configurations.",

		ReadContext: dataUserRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Username of the user",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"invitation_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email address used for the user invitation",
			},
			"policy": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of policies (space access rules) assigned to the user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID (slug) of the space the user has access to",
							Computed:    true,
						},
						"role": {
							Type: schema.TypeString,
							Description: "Type of access to the space. Possible values are: " +
								"READ, WRITE, ADMIN",
							Computed: true,
						},
					},
				},
				Set: userPolicyHash,
			},
		},
	}
}

func dataUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	username := d.Get("username").(string)

	var query struct {
		ManagedUsers []structs.User `graphql:"managedUsers"`
	}

	if err := meta.(*internal.Client).Query(ctx, "ManagedUsersRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for users: %v", err)
	}

	var user *structs.User
	for _, u := range query.ManagedUsers {
		if u.Username == username {
			user = &u
			break
		}
	}

	if user == nil {
		return diag.Errorf("user with username %q not found", username)
	}

	d.SetId(user.ID)
	d.Set("username", user.Username)
	d.Set("invitation_email", user.InvitationEmail)

	var accessList []interface{}
	for _, a := range user.AccessRules {
		accessList = append(accessList, map[string]interface{}{
			"space_id": a.Space,
			"role":     a.SpaceAccessLevel,
		})
	}
	d.Set("policy", accessList)

	return nil
}
