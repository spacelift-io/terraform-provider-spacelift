package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_api_key` represents a Spacelift API Key - " +
			"a credential that can be used to authenticate with the Spacelift API " +
			"from outside of Spacelift, typically for automation purposes.",

		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the API key",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"idp_groups": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of idp groups associated with the API key",
				Optional:    true,
				Set:         schema.HashString,
			},
			"oidc": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Description: "OIDC configuration for the API key. When provided, creates an OIDC API key instead of a SECRET API key.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer": {
							Type:             schema.TypeString,
							Description:      "OIDC issuer URL",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"client_id": {
							Type:             schema.TypeString,
							Description:      "OIDC client ID",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"subject_expression": {
							Type:             schema.TypeString,
							Description:      "OIDC subject expression",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
					},
				},
			},
			"access_rule": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of space access rules for the API key",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:             schema.TypeString,
							Description:      "ID of the space this access rule applies to",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "Role for the space. Can be built-in access levels (READ, WRITE, ADMIN) or a custom role name",
							Required:    true,
						},
					},
				},
				Set: accessRuleHash,
			},
			"secret": {
				Type:        schema.TypeString,
				Description: "The secret value of the API key",
				Computed:    true,
				Sensitive:   true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the API key (SECRET or OIDC)",
				Computed:    true,
			},
		},
	}
}

func accessRuleHash(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	spaceID, _ := m["space_id"].(string)
	accessLevel, _ := m["role"].(string)

	key := spaceID + "-" + accessLevel
	return schema.HashString(key)
}

func apiKeyCreateInput(d *schema.ResourceData) structs.ApiKeyInput {
	input := structs.ApiKeyInput{
		Name:  graphql.String(d.Get("name").(string)),
		Admin: graphql.Boolean(false), // Always false - we don't use this field
	}

	// Always set IDPGroups to ensure we send an empty array instead of null
	// Initialize as an empty slice (not nil) to ensure JSON serialization sends []
	idpGroups := make([]graphql.String, 0)
	if idpGroupsSet, ok := d.Get("idp_groups").(*schema.Set); ok && idpGroupsSet != nil {
		for _, team := range idpGroupsSet.List() {
			idpGroups = append(idpGroups, graphql.String(team.(string)))
		}
	}
	input.IDPGroups = idpGroups

	// Add OIDC configuration if provided
	if oidcList, ok := d.Get("oidc").([]interface{}); ok && len(oidcList) > 0 {
		if oidcMap, ok := oidcList[0].(map[string]interface{}); ok {
			input.OIDC = &structs.APIKeyInputOIDC{
				Issuer:            graphql.String(oidcMap["issuer"].(string)),
				ClientID:          graphql.String(oidcMap["client_id"].(string)),
				SubjectExpression: graphql.String(oidcMap["subject_expression"].(string)),
			}
		}
	}

	// Don't include AccessRules in the input - we'll handle them separately with role bindings

	return input
}

// getRoleIDForAccessLevel queries the available roles and returns the role ID for the given access level or role name
// It handles both built-in access levels (READ/WRITE/ADMIN) and custom role names
func getRoleIDForAccessLevel(ctx context.Context, client *internal.Client, accessLevel string) (string, error) {
	// Query all available roles
	var query struct {
		Roles []struct {
			ID   string `graphql:"id"`
			Name string `graphql:"name"`
		} `graphql:"roles"`
	}

	if err := client.Query(ctx, "GetRoles", &query, map[string]interface{}{}); err != nil {
		return "", fmt.Errorf("could not query roles: %v", err)
	}

	// First check if it's a built-in access level that needs to be mapped to a role name
	var targetRoleName string
	switch accessLevel {
	case "READ":
		targetRoleName = "Space reader"
	case "WRITE":
		targetRoleName = "Space writer"
	case "ADMIN":
		targetRoleName = "Space admin"
	default:
		// For anything else, treat it as a custom role name and search for it directly
		targetRoleName = accessLevel
	}

	// Find the role ID for the target role name (built-in or custom)
	for _, role := range query.Roles {
		if role.Name == targetRoleName {
			return role.ID, nil
		}
	}

	return "", fmt.Errorf("role not found: %s", accessLevel)
}

// manageAPIKeyRoleBindings handles creating/updating role bindings for an API key
func manageAPIKeyRoleBindings(ctx context.Context, client *internal.Client, apiKeyID string, accessRules []interface{}) error {
	if len(accessRules) == 0 {
		return nil // No access rules to manage
	}

	var bindings []structs.ApiKeyRoleBindingInput
	for _, rule := range accessRules {
		ruleMap := rule.(map[string]interface{})
		spaceID := ruleMap["space_id"].(string)
		accessLevel := ruleMap["role"].(string)

		roleID, err := getRoleIDForAccessLevel(ctx, client, accessLevel)
		if err != nil {
			return fmt.Errorf("could not get role ID for access level %s: %v", accessLevel, err)
		}

		bindings = append(bindings, structs.ApiKeyRoleBindingInput{
			APIKeyID: graphql.ID(apiKeyID),
			RoleID:   graphql.ID(roleID),
			SpaceID:  graphql.ID(spaceID),
		})
	}

	// Create role bindings using batch mutation
	var mutation struct {
		RoleBindings []structs.APIKeyRoleBinding `graphql:"apiKeyRoleBindingBatchCreate(input: $input)"`
	}

	input := structs.ApiKeyRoleBindingBatchInput{
		Bindings: bindings,
	}

	variables := map[string]interface{}{
		"input": input,
	}

	if err := client.Mutate(ctx, "APIKeyRoleBindingBatchCreate", &mutation, variables); err != nil {
		return fmt.Errorf("could not create role bindings: %v", err)
	}

	return nil
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*internal.Client)

	// Create the API key first
	var mutation struct {
		APIKey structs.APIKey `graphql:"apiKeyCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": apiKeyCreateInput(d),
	}

	if err := client.Mutate(ctx, "APIKeyCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create API key %v: %v", d.Get("name").(string), err)
	}

	apiKeyID := mutation.APIKey.ID
	d.SetId(apiKeyID)

	// Capture the secret immediately - it's only available once during creation
	if mutation.APIKey.Secret != "" {
		d.Set("secret", mutation.APIKey.Secret)
	}

	// Also set other fields that are available from the creation response
	d.Set("name", mutation.APIKey.Name)
	d.Set("type", string(mutation.APIKey.Type))

	// Handle access rules using role bindings
	if accessRulesSet, ok := d.Get("access_rule").(*schema.Set); ok && accessRulesSet.Len() > 0 {
		if err := manageAPIKeyRoleBindings(ctx, client, apiKeyID, accessRulesSet.List()); err != nil {
			// If role bindings fail, we should delete the API key to avoid leaving it in an incomplete state
			deleteErr := deleteAPIKey(ctx, client, apiKeyID)
			if deleteErr != nil {
				return diag.Errorf("could not create role bindings for API key %v: %v (also failed to clean up API key: %v)", d.Get("name").(string), err, deleteErr)
			}
			return diag.Errorf("could not create role bindings for API key %v: %v", d.Get("name").(string), err)
		}
	}

	return resourceAPIKeyRead(ctx, d, meta)
}

// deleteAPIKey is a helper function to delete an API key by ID
func deleteAPIKey(ctx context.Context, client *internal.Client, apiKeyID string) error {
	var mutation struct {
		APIKey *structs.APIKey `graphql:"apiKeyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(apiKeyID)}
	return client.Mutate(ctx, "APIKeyDelete", &mutation, variables)
}

// getCurrentRoleBindings gets existing role bindings for an API key
func getCurrentRoleBindings(ctx context.Context, client *internal.Client, apiKeyID string) ([]structs.APIKeyRoleBinding, error) {
	var query struct {
		APIKey *struct {
			RoleBindings []structs.APIKeyRoleBinding `graphql:"apiKeyRoleBindings"`
		} `graphql:"apiKey(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(apiKeyID)}
	if err := client.Query(ctx, "GetAPIKeyRoleBindings", &query, variables); err != nil {
		return nil, err
	}

	if query.APIKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	return query.APIKey.RoleBindings, nil
}

// deleteRoleBindings deletes specific role bindings by their IDs
func deleteRoleBindings(ctx context.Context, client *internal.Client, roleBindingIDs []string) error {
	for _, bindingID := range roleBindingIDs {
		var mutation struct {
			RoleBinding *structs.APIKeyRoleBinding `graphql:"apiKeyRoleBindingDelete(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.ID(bindingID)}
		if err := client.Mutate(ctx, "APIKeyRoleBindingDelete", &mutation, variables); err != nil {
			return fmt.Errorf("could not delete role binding %s: %v", bindingID, err)
		}
	}
	return nil
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		APIKey *structs.APIKey `graphql:"apiKey(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "APIKeyRead", &query, variables); err != nil {
		return diag.Errorf("could not query for API key: %v", err)
	}

	apiKey := query.APIKey
	if apiKey == nil {
		d.SetId("")
		return nil
	}

	d.SetId(apiKey.ID)
	d.Set("name", apiKey.Name)
	d.Set("type", string(apiKey.Type))

	idpGroups := schema.NewSet(schema.HashString, []interface{}{})
	for _, team := range apiKey.IDPGroups {
		idpGroups.Add(team)
	}
	d.Set("idp_groups", idpGroups)

	accessRules := schema.NewSet(accessRuleHash, []interface{}{})

	// Use role bindings created via apiKeyRoleBindingBatchCreate
	for _, binding := range apiKey.RoleBindings {
		// Convert role name back to the format expected by the user
		var roleName string
		switch binding.Role.Name {
		case "Space reader":
			roleName = "READ"
		case "Space writer":
			roleName = "WRITE"
		case "Space admin":
			roleName = "ADMIN"
		default:
			// Custom role - use the actual role name
			roleName = binding.Role.Name
		}

		ruleMap := map[string]interface{}{
			"space_id": binding.SpaceID,
			"role":     roleName,
		}
		accessRules.Add(ruleMap)
	}

	// Fallback to old AccessRules field in case they exist (for backward compatibility)
	for _, rule := range apiKey.AccessRules {
		ruleMap := map[string]interface{}{
			"space_id": rule.Space,
			"role":     rule.SpaceAccessLevel,
		}
		accessRules.Add(ruleMap)
	}

	d.Set("access_rule", accessRules)

	if apiKey.Secret != "" {
		d.Set("secret", apiKey.Secret)
	}

	return nil
}

func apiKeyUpdateInput(d *schema.ResourceData) structs.ApiKeyUpdateInput {
	input := structs.ApiKeyUpdateInput{}

	if d.HasChange("name") {
		name := graphql.String(d.Get("name").(string))
		input.Name = &name
	}

	if d.HasChange("idp_groups") {
		if idpGroupsSet, ok := d.Get("idp_groups").(*schema.Set); ok {
			var idpGroups []graphql.String
			for _, group := range idpGroupsSet.List() {
				idpGroups = append(idpGroups, graphql.String(group.(string)))
			}
			input.IDPGroups = idpGroups
		}
	}

	return input
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*internal.Client)
	apiKeyID := d.Id()

	// Update basic fields (name, idp_groups) if they changed
	if d.HasChange("name") || d.HasChange("idp_groups") {
		var mutation struct {
			APIKey structs.APIKey `graphql:"apiKeyUpdate(id: $id, input: $input)"`
		}

		variables := map[string]interface{}{
			"id":    graphql.ID(apiKeyID),
			"input": apiKeyUpdateInput(d),
		}

		if err := client.Mutate(ctx, "APIKeyUpdate", &mutation, variables); err != nil {
			return diag.Errorf("could not update API key: %v", err)
		}
	}

	// Handle access rules changes using role bindings
	if d.HasChange("access_rule") {
		// Get current role bindings
		currentBindings, err := getCurrentRoleBindings(ctx, client, apiKeyID)
		if err != nil {
			return diag.Errorf("could not get current role bindings: %v", err)
		}

		// Delete all existing role bindings
		var bindingIDs []string
		for _, binding := range currentBindings {
			bindingIDs = append(bindingIDs, binding.ID)
		}

		if len(bindingIDs) > 0 {
			if err := deleteRoleBindings(ctx, client, bindingIDs); err != nil {
				return diag.Errorf("could not delete existing role bindings: %v", err)
			}
		}

		// Create new role bindings based on the updated access rules
		if accessRulesSet, ok := d.Get("access_rule").(*schema.Set); ok && accessRulesSet.Len() > 0 {
			if err := manageAPIKeyRoleBindings(ctx, client, apiKeyID, accessRulesSet.List()); err != nil {
				return diag.Errorf("could not create new role bindings: %v", err)
			}
		}
	}

	return resourceAPIKeyRead(ctx, d, meta)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		APIKey *structs.APIKey `graphql:"apiKeyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "APIKeyDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete API key: %v", err)
	}

	d.SetId("")

	return nil
}
