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

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_user` represents a mapping between a Spacelift user " +
			"(managed using an Identity Provider) and a Policy. A Policy defines " +
			"what access rights the user has to a given Space.",
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
				Required:    true,
				Description: "Username of the user",
			},
			"policy": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:             schema.TypeString,
							Description:      "ID (slug) of the space the user has access to",
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
			"invitation_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "`invitation_email` will be used to send an invitation to the specified email address. This property is required when creating a new user. This property is optional when importing an existing user.",
			},
		},
	}
}

func userPolicyHash(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	spaceID, _ := m["space_id"].(string)
	role, _ := m["role"].(string)

	key := spaceID + "-" + role
	return schema.HashString(key)
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	// send an Invite (create) mutation to the API
	var mutation struct {
		User *structs.User `graphql:"managedUserInvite(input: $input)"`
	}

	var email *graphql.String
	if d.Get("invitation_email") != "" {
		email = toOptionalString(d.Get("invitation_email"))
	}

	if email == nil || *email == "" {
		return diag.Errorf("invitation_email is required for new users")
	}

	variables := map[string]interface{}{
		"input": structs.ManagedUserInviteInput{
			InvitationEmail: email,
			Username:        toString(d.Get("username")),
			AccessRules:     getAccessRules(d),
		},
	}
	if err := i.(*internal.Client).Mutate(ctx, "ManagedUserInvite", &mutation, variables); err != nil {
		return diag.Errorf("could not create user mapping %s: %v", toString(d.Get("username")), internal.FromSpaceliftError(err))
	}

	// set the ID in TF state
	d.SetId(mutation.User.ID)

	// fetch state from remote and write to TF state
	return resourceUserRead(ctx, d, i)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	// send a read query to the API
	var query struct {
		User *structs.User `graphql:"managedUser(id: $id)"`
	}
	variables := map[string]interface{}{"id": toID(d.Id())}
	if err := i.(*internal.Client).Query(ctx, "ManagedUser", &query, variables); err != nil {
		return diag.Errorf("could not query for user mapping: %v", err)
	}

	// if the mapping is not found on the remote side, delete it from the TF state
	if query.User == nil {
		d.SetId("")
		return nil
	}

	// if found, update the TF state
	d.Set("invitation_email", query.User.InvitationEmail)
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
	// input validation
	if d.HasChange("invitation_email") {
		return diag.Errorf("invitation_email cannot be changed")
	}
	if d.HasChange("username") {
		return diag.Errorf("username cannot be changed")
	}

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
		ret = append(ret, diag.Errorf("could not update user mapping %s: %v", d.Id(), internal.FromSpaceliftError(err))...)
	}

	// fetch from remote and write to TF state
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
		return diag.Errorf("could not delete user mapping %s: %v", d.Id(), internal.FromSpaceliftError(err))
	}

	// if the user was deleted, remove it from the TF state as well
	d.SetId("")

	return nil
}
