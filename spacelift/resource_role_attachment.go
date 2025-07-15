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
)

func resourceRoleAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_role_attachment` represents a Spacelift role attachment " +
			"between an API key and a role; or between an IdP Group Mapping and a role.\n" +
			"Either `api_key_id` or `idp_group_mapping_id` must be set, but not both.",

		CreateContext: resourceRoleAttachmentCreate,
		ReadContext:   resourceRoleAttachmentRead,
		DeleteContext: resourceRoleAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_key_id": {
				Type:         schema.TypeString,
				Description:  "ID of the API key to attach to the role (ULID format). For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"api_key_id", "idp_group_mapping_id"},
			},
			"idp_group_mapping_id": {
				Type:        schema.TypeString,
				Description: "ID of the IdP Group Mapping to attach to the role (ULID format). For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
				Optional:    true,
				ForceNew:    true,
			},
			"role_id": {
				Type:             schema.TypeString,
				Description:      "ID of the role to attach to the API key (ULID format). For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.",
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

	if apiKeyID != "" {
		return createAPIKeyRoleBinding(ctx, d, meta)
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

func resourceRoleAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	if strings.HasPrefix(id, apiKeyRoleAttachmentPrefix) {
		return readAPIKeyRoleBinding(ctx, d, meta)

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
	d.Set("role_id", roleBinding.Role.ID)
	d.Set("space_id", roleBinding.SpaceID)

	return nil
}

func resourceRoleAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	if strings.HasPrefix(id, apiKeyRoleAttachmentPrefix) {
		return deleteAPIKeyRoleBinding(ctx, d, meta)
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
