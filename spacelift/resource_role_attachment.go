package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

const (
	apiKeyRoleAttachmentPrefix          = "API"
	idpGroupMappingRoleAttachmentPrefix = "IDP"
	stackRoleAttachmentPrefix           = "STACK"
	userRoleAttachmentPrefix            = "USER"
)

func resourceRoleAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_role_attachment` represents a Spacelift role attachment " +
			"between an API key and a role within a specific space, " +
			"an IdP Group Mapping and a role within a specific space, " +
			"a stack and a role within a specific space, " +
			"or a user and a role within a specific space. " +
			"Exactly one of `api_key_id`, `idp_group_mapping_id`, `stack_id`, or `user_id` must be set.",

		CreateContext: resourceRoleAttachmentCreate,
		ReadContext:   resourceRoleAttachmentRead,
		DeleteContext: resourceRoleAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_key_id": {
				Type:         schema.TypeString,
				Description:  "ID of the API key (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"api_key_id", "idp_group_mapping_id", "stack_id", "user_id"},
			},
			"idp_group_mapping_id": {
				Type:         schema.TypeString,
				Description:  "ID of the IdP Group Mapping (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"api_key_id", "idp_group_mapping_id", "stack_id", "user_id"},
			},
			"stack_id": {
				Type:         schema.TypeString,
				Description:  "Slug of the Stack to attach to the role. For example: `my-stack`.",
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"api_key_id", "idp_group_mapping_id", "stack_id", "user_id"},
			},
			"user_id": {
				Type:         schema.TypeString,
				Description:  "ID of the user (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"api_key_id", "idp_group_mapping_id", "stack_id", "user_id"},
			},
			"role_id": {
				Type:             schema.TypeString,
				Description:      "ID of the role (ULID format) to attach to the API key, IdP Group, stack, or user. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space_id": {
				Type:             schema.TypeString,
				Description:      "ID of the space where the role attachment should be created",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
		},
	}
}

func resourceRoleAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiKeyID := d.Get("api_key_id").(string)
	stackID := d.Get("stack_id").(string)
	userID := d.Get("user_id").(string)

	if apiKeyID != "" {
		return createAPIKeyRoleBinding(ctx, d, meta)
	}

	if stackID != "" {
		return createStackRoleBinding(ctx, d, meta)
	}

	if userID != "" {
		return createUserRoleBinding(ctx, d, meta)
	}

	return createIDPGroupMappingRoleBinding(ctx, d, meta)
}

func createAPIKeyRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateRoleBinding structs.APIKeyRoleBinding `graphql:"apiKeyRoleBindingCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.ApiKeyRoleBindingInput{
			APIKeyID: toID(d.Get("api_key_id").(string)),
			RoleID:   toID(d.Get("role_id").(string)),
			SpaceID:  toID(d.Get("space_id").(string)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ApiKeyRoleBindingCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create role attachment: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", apiKeyRoleAttachmentPrefix, mutation.CreateRoleBinding.ID))

	return resourceRoleAttachmentRead(ctx, d, meta)
}

func createUserRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateRoleBindings []structs.UserRoleBinding `graphql:"userRoleBindingBatchCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.UserRoleBindingBatchInput{
			Bindings: []structs.UserRoleBindingInput{
				{
					UserID:  toID(d.Get("user_id").(string)),
					RoleID:  toID(d.Get("role_id").(string)),
					SpaceID: toID(d.Get("space_id").(string)),
				},
			},
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "UserRoleBindingBatchCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create user role binding: %v", internal.FromSpaceliftError(err))
	}

	if len(mutation.CreateRoleBindings) == 0 {
		return diag.Errorf("no user role binding was created")
	}

	d.SetId(fmt.Sprintf("%s/%s", userRoleAttachmentPrefix, mutation.CreateRoleBindings[0].ID))

	return resourceRoleAttachmentRead(ctx, d, meta)
}

func createIDPGroupMappingRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateRoleBinding structs.UserGroupRoleBinding `graphql:"userGroupRoleBindingCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.UserGroupRoleBindingInput{
			UserGroupID: toID(d.Get("idp_group_mapping_id").(string)),
			RoleID:      toID(d.Get("role_id").(string)),
			SpaceID:     toID(d.Get("space_id").(string)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "UserGroupRoleBindingCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create user group role binding: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", idpGroupMappingRoleAttachmentPrefix, mutation.CreateRoleBinding.ID))

	return resourceRoleAttachmentRead(ctx, d, meta)
}

func createStackRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateRoleBinding structs.StackRoleBinding `graphql:"stackRoleBindingCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.StackRoleBindingInput{
			StackID: toID(d.Get("stack_id").(string)),
			RoleID:  toID(d.Get("role_id").(string)),
			SpaceID: toID(d.Get("space_id").(string)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackRoleBindingCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create stack role binding: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", stackRoleAttachmentPrefix, mutation.CreateRoleBinding.ID))

	return resourceRoleAttachmentRead(ctx, d, meta)
}

func resourceRoleAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	if strings.HasPrefix(id, apiKeyRoleAttachmentPrefix) {
		return readAPIKeyRoleBinding(ctx, d, meta)
	}

	if strings.HasPrefix(id, stackRoleAttachmentPrefix) {
		return readStackRoleBinding(ctx, d, meta)
	}

	if strings.HasPrefix(id, userRoleAttachmentPrefix) {
		return readUserRoleBinding(ctx, d, meta)
	}

	return readIDPGroupMappingRoleBinding(ctx, d, meta)
}

func readAPIKeyRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		APIKeyRoleBinding *structs.APIKeyRoleBinding `graphql:"apiKeyRoleBinding(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), apiKeyRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Query(ctx, "ApiKeyRoleBindingRead", &query, variables); err != nil {
		return diag.Errorf("could not query for role attachment: %v", internal.FromSpaceliftError(err))
	}

	if query.APIKeyRoleBinding == nil {
		d.SetId("")
		return nil
	}

	roleBinding := query.APIKeyRoleBinding

	d.Set("api_key_id", roleBinding.APIKeyID)
	d.Set("role_id", roleBinding.Role.ID)
	d.Set("space_id", roleBinding.SpaceID)

	return nil
}

func readUserRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		UserRoleBinding *structs.UserRoleBinding `graphql:"userRoleBinding(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), userRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Query(ctx, "UserRoleBindingRead", &query, variables); err != nil {
		return diag.Errorf("could not query for user role binding: %v", internal.FromSpaceliftError(err))
	}

	if query.UserRoleBinding == nil {
		d.SetId("")
		return nil
	}

	roleBinding := query.UserRoleBinding

	d.Set("user_id", roleBinding.UserID)
	d.Set("role_id", roleBinding.RoleID)
	d.Set("space_id", roleBinding.SpaceID)

	return nil
}

func readIDPGroupMappingRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		UserGroupRoleBinding *structs.UserGroupRoleBinding `graphql:"userGroupRoleBinding(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), idpGroupMappingRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Query(ctx, "UserGroupRoleBindingRead", &query, variables); err != nil {
		return diag.Errorf("could not query for user group role binding: %v", internal.FromSpaceliftError(err))
	}

	if query.UserGroupRoleBinding == nil {
		d.SetId("")
		return nil
	}

	roleBinding := query.UserGroupRoleBinding

	d.Set("idp_group_mapping_id", roleBinding.UserGroup.ID)
	d.Set("role_id", roleBinding.RoleID)
	d.Set("space_id", roleBinding.SpaceID)

	return nil
}

func readStackRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackRoleBinding *structs.StackRoleBinding `graphql:"stackRoleBinding(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), stackRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Query(ctx, "StackRoleBindingRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack role binding: %v", internal.FromSpaceliftError(err))
	}

	if query.StackRoleBinding == nil {
		d.SetId("")
		return nil
	}

	roleBinding := query.StackRoleBinding

	d.Set("stack_id", roleBinding.StackID)
	d.Set("role_id", roleBinding.RoleID)
	d.Set("space_id", roleBinding.SpaceID)

	return nil
}

func resourceRoleAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	if strings.HasPrefix(id, apiKeyRoleAttachmentPrefix) {
		return deleteAPIKeyRoleBinding(ctx, d, meta)
	}

	if strings.HasPrefix(id, stackRoleAttachmentPrefix) {
		return deleteStackRoleBinding(ctx, d, meta)
	}

	if strings.HasPrefix(id, userRoleAttachmentPrefix) {
		return deleteUserRoleBinding(ctx, d, meta)
	}

	return deleteIDPGroupMappingRoleBinding(ctx, d, meta)
}

func deleteAPIKeyRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteRoleBinding *structs.APIKeyRoleBinding `graphql:"apiKeyRoleBindingDelete(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), apiKeyRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ApiKeyRoleBindingDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete role attachment: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func deleteUserRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteRoleBinding *structs.UserRoleBinding `graphql:"userRoleBindingDelete(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), userRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "UserRoleBindingDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete user role binding: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func deleteIDPGroupMappingRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteRoleBinding *structs.UserGroupRoleBinding `graphql:"userGroupRoleBindingDelete(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), idpGroupMappingRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "UserGroupRoleBindingDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete user group role binding: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func deleteStackRoleBinding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteRoleBinding *structs.StackRoleBinding `graphql:"stackRoleBindingDelete(id: $id)"`
	}

	id := strings.TrimPrefix(d.Id(), stackRoleAttachmentPrefix+"/")
	variables := map[string]interface{}{
		"id": toID(id),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackRoleBindingDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack role binding: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
