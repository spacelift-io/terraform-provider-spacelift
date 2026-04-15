package spacelift

import (
	"context"

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
			"from outside of Spacelift, typically for automation purposes.\n\n" +
			"### WARNING\n\n" +
			"**This resource manages API keys which are sensitive credentials. " +
			"These keys will be saved to your state file. Ensure that you handle your state file securely.**",

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
							ForceNew:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"client_id": {
							Type:             schema.TypeString,
							Description:      "OIDC client ID",
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"subject_expression": {
							Type:             schema.TypeString,
							Description:      "OIDC subject expression",
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"claim_mappings": {
							Type:        schema.TypeMap,
							Description: "Mapping of IdP JWT claim names to Spacelift claim names. Currently only the `groups` Spacelift claim is supported, so the value should always be `groups` (e.g. `teams = \"groups\"` maps the IdP `teams` claim to Spacelift `groups`). When configured, the mapped claims are extracted from the OIDC token at authentication time, enabling dynamic team membership based on the authenticating user's actual IdP groups.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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

	if oidcList, ok := d.Get("oidc").([]any); ok && len(oidcList) > 0 {
		if oidcMap, ok := oidcList[0].(map[string]any); ok {
			oidcInput := &structs.APIKeyInputOIDC{
				Issuer:            graphql.String(oidcMap["issuer"].(string)),
				ClientID:          graphql.String(oidcMap["client_id"].(string)),
				SubjectExpression: graphql.String(oidcMap["subject_expression"].(string)),
			}

			if cm, ok := oidcMap["claim_mappings"].(map[string]any); ok && len(cm) > 0 {
				oidcInput.ClaimMappings = toClaimMappingInput(cm)
			}

			input.OIDC = oidcInput
		}
	}

	return input
}

func toClaimMappingInput(cm map[string]any) *structs.ClaimMappingInput {
	entries := make([]structs.ClaimMappingEntryInput, 0, len(cm))
	for name, value := range cm {
		entries = append(entries, structs.ClaimMappingEntryInput{
			Name:  graphql.String(name),
			Value: graphql.String(value.(string)),
		})
	}
	return &structs.ClaimMappingInput{Entries: entries}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*internal.Client)

	// Create the API key first
	var mutation struct {
		APIKey structs.APIKey `graphql:"apiKeyCreate(input: $input)"`
	}

	variables := map[string]any{
		"input": apiKeyCreateInput(d),
	}

	if err := client.Mutate(ctx, "APIKeyCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create API key %v: %v", d.Get("name").(string), err)
	}

	apiKeyID := mutation.APIKey.ID
	d.SetId(apiKeyID)
	d.Set("secret", mutation.APIKey.Secret)
	d.Set("name", mutation.APIKey.Name)
	d.Set("type", string(mutation.APIKey.Type))

	return resourceAPIKeyRead(ctx, d, meta)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		APIKey *structs.APIKey `graphql:"apiKey(id: $id)"`
	}

	variables := map[string]any{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "APIKeyRead", &query, variables); err != nil {
		if err.Error() == "could not find api key" {
			d.SetId("")
			return nil
		}
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

	idpGroups := schema.NewSet(schema.HashString, []any{})
	for _, team := range apiKey.IDPGroups {
		idpGroups.Add(team)
	}
	d.Set("idp_groups", idpGroups)

	if apiKey.OIDCSettings != nil {
		oidc := map[string]any{
			"issuer":             apiKey.OIDCSettings.Issuer,
			"client_id":          apiKey.OIDCSettings.ClientID,
			"subject_expression": apiKey.OIDCSettings.SubjectExpression,
		}

		claimMappings := map[string]string{}
		for _, entry := range apiKey.OIDCSettings.ClaimMapping.Entries {
			claimMappings[entry.Name] = entry.Value
		}
		if len(claimMappings) > 0 {
			oidc["claim_mappings"] = claimMappings
		}

		d.Set("oidc", []any{oidc})
	}

	return nil
}

func apiKeyUpdateInput(d *schema.ResourceData) structs.ApiKeyUpdateInput {
	input := structs.ApiKeyUpdateInput{}

	if d.HasChange("name") {
		name := graphql.String(d.Get("name").(string))
		input.Name = &name
	}

	if idpGroupsSet, ok := d.Get("idp_groups").(*schema.Set); ok {
		var idpGroups []graphql.String
		for _, group := range idpGroupsSet.List() {
			idpGroups = append(idpGroups, graphql.String(group.(string)))
		}
		input.IDPGroups = idpGroups
	}

	if d.HasChange("oidc.0.claim_mappings") {
		if oidcList, ok := d.Get("oidc").([]any); ok && len(oidcList) > 0 {
			if oidcMap, ok := oidcList[0].(map[string]any); ok {
				if cm, ok := oidcMap["claim_mappings"].(map[string]any); ok {
					input.OIDCClaimMappings = toClaimMappingInput(cm)
				} else {
					input.OIDCClaimMappings = &structs.ClaimMappingInput{Entries: []structs.ClaimMappingEntryInput{}}
				}
			}
		}
	}

	return input
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*internal.Client)
	apiKeyID := d.Id()

	// Update basic fields (name, idp_groups) if they changed
	if d.HasChange("name") || d.HasChange("idp_groups") || d.HasChange("oidc.0.claim_mappings") {
		var mutation struct {
			APIKey structs.APIKey `graphql:"apiKeyUpdate(id: $id, input: $input)"`
		}

		variables := map[string]any{
			"id":    graphql.ID(apiKeyID),
			"input": apiKeyUpdateInput(d),
		}

		if err := client.Mutate(ctx, "APIKeyUpdate", &mutation, variables); err != nil {
			return diag.Errorf("could not update API key: %v", err)
		}
	}

	return resourceAPIKeyRead(ctx, d, meta)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		APIKey *structs.APIKey `graphql:"apiKeyDelete(id: $id)"`
	}

	variables := map[string]any{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "APIKeyDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete API key: %v", err)
	}

	d.SetId("")

	return nil
}
