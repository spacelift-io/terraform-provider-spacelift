package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAIIntegrations() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_ai_integrations` represents a list of all the AI integrations in the Spacelift " +
			"account visible to the API user, including the read-only ones provided by Spacelift.",

		ReadContext: dataAIIntegrationsRead,

		Schema: map[string]*schema.Schema{
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "required labels to match",
				Optional:    true,
			},
			"ai_provider": {
				Type:             schema.TypeString,
				Description:      "Only return the integrations backed by this AI provider, one of `Anthropic`, `Bedrock`, `Google` or `OpenAI`",
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(aiProviderNames, false)),
			},
			"spacelift_provided": {
				Type:        schema.TypeBool,
				Description: "Only return the integrations that are, or are not, provided by Spacelift. Leave unset to return both.",
				Optional:    true,
			},
			"integrations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_id": {
							Type:        schema.TypeString,
							Description: "Immutable ID of the integration",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Friendly name of the integration",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Free-form description of the integration",
							Computed:    true,
						},
						"ai_provider": {
							Type:        schema.TypeString,
							Description: "AI provider backing the integration, one of `Anthropic`, `OpenAI`, `Google` or `Bedrock`",
							Computed:    true,
						},
						"models": {
							Type:        schema.TypeList,
							Description: "Model identifiers available on this integration",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
						"labels": {
							Type:        schema.TypeSet,
							Description: "Labels set on the integration",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
						"space_id": {
							Type:        schema.TypeString,
							Description: "ID of the space the integration belongs to, empty for Spacelift-provided integrations",
							Computed:    true,
						},
						"base_url": {
							Type:        schema.TypeString,
							Description: "Custom base URL of the provider-compatible gateway the integration calls",
							Computed:    true,
						},
						"provider_integration_id": {
							Type:        schema.TypeString,
							Description: "ID of the underlying provider integration, for example the AWS integration backing a Bedrock integration",
							Computed:    true,
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Whether the integration can be used for LLM calls",
							Computed:    true,
						},
						"is_spacelift_provided": {
							Type:        schema.TypeBool,
							Description: "Whether this is a read-only, Spacelift-managed integration shared with every account",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataAIIntegrationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		AIIntegrations []*structs.AIIntegration `graphql:"aiIntegrations()"`
	}

	if err := meta.(*internal.Client).Query(ctx, "AIIntegrationsRead", &query, map[string]any{}); err != nil {
		return diag.Errorf("could not query for AI integrations: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("spacelift_ai_integrations")

	if query.AIIntegrations == nil {
		d.Set("integrations", nil)
		return nil
	}

	filtered := internal.FilterByRequiredLabels(d, query.AIIntegrations, func(integration *structs.AIIntegration) []string {
		return integration.Labels
	})

	provider, byProvider := d.GetOk("ai_provider")
	// d.Get cannot tell an unset boolean from false, so the raw config decides
	// whether the filter was asked for at all.
	spaceliftProvided := getOptionalBool(d, "spacelift_provided")

	mapped := make([]map[string]any, 0, len(filtered))
	for _, integration := range filtered {
		if byProvider && integration.Provider != provider.(string) {
			continue
		}

		if spaceliftProvided != nil && integration.IsSpaceliftProvided != bool(*spaceliftProvided) {
			continue
		}

		flattened := flattenAIIntegration(integration)
		flattened["integration_id"] = integration.ID
		mapped = append(mapped, flattened)
	}

	if err := d.Set("integrations", mapped); err != nil {
		d.SetId("")
		return diag.Errorf("could not set AI integrations: %v", err)
	}

	return nil
}

// flattenAIIntegration maps an integration onto the attributes shared by both
// data sources. The list keeps provider-specific values flat, the way
// spacelift_stacks exposes a plain vendor field.
func flattenAIIntegration(integration *structs.AIIntegration) map[string]any {
	labels := schema.NewSet(schema.HashString, []any{})
	for _, label := range integration.Labels {
		labels.Add(label)
	}

	mapped := map[string]any{
		"name":                    integration.Name,
		"description":             "",
		"ai_provider":             integration.Provider,
		"models":                  integration.Models,
		"labels":                  labels,
		"space_id":                "",
		"base_url":                "",
		"provider_integration_id": "",
		"enabled":                 integration.Enabled,
		"is_spacelift_provided":   integration.IsSpaceliftProvided,
	}

	if integration.Description != nil {
		mapped["description"] = *integration.Description
	}

	if integration.BaseURL != nil {
		mapped["base_url"] = *integration.BaseURL
	}

	if integration.ProviderIntegrationID != nil {
		mapped["provider_integration_id"] = *integration.ProviderIntegrationID
	}

	if integration.Space != nil {
		mapped["space_id"] = integration.Space.ID
	}

	return mapped
}
