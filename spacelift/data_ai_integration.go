package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataAIIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_ai_integration` represents an integration with an AI/LLM provider.",

		ReadContext: dataAIIntegrationRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:             schema.TypeString,
				Description:      "Immutable ID of the integration",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
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
			"anthropic": dataAIProviderBlock("Anthropic", "Anthropic"),
			"google":    dataAIProviderBlock("Google", "Gemini"),
			"openai":    dataAIProviderBlock("OpenAI", "OpenAI"),
			"bedrock": {
				Type: schema.TypeList,
				Description: "AWS Bedrock-specific configuration. Presence means this integration calls " +
					"Bedrock through an AWS integration rather than using an API key.",
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_id": {
							Type:        schema.TypeString,
							Description: "ID of the AWS integration used to reach Bedrock",
							Computed:    true,
						},
						"integration_name": {
							Type:        schema.TypeString,
							Description: "Name of the AWS integration used to reach Bedrock",
							Computed:    true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "AWS region the Bedrock requests are routed through",
							Computed:    true,
						},
						"profiles": {
							Type:        schema.TypeList,
							Description: "Bedrock inference profile ARNs exposed by this integration",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
					},
				},
			},
			"models": {
				Type: schema.TypeList,
				Description: "Model identifiers this integration is pinned to, empty when it follows the " +
					"default list for its provider. For Bedrock, its inference profiles.",
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
	}
}

// dataAIProviderBlock mirrors the resource's blocks. The API key is never
// readable, so it is absent here.
func dataAIProviderBlock(label, api string) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Description: fmt.Sprintf(
			"%s-specific configuration. Presence means this integration uses the %s API.",
			label, api,
		),
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"base_url": {
					Type: schema.TypeString,
					Description: fmt.Sprintf(
						"Custom base URL of the %s-compatible gateway the integration calls", api,
					),
					Computed: true,
				},
			},
		},
	}
}

func dataAIIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		AIIntegration *structs.AIIntegration `graphql:"aiIntegration(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Get("integration_id").(string))}
	if err := meta.(*internal.Client).Query(ctx, "AIIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for AI integration: %v", internal.FromSpaceliftError(err))
	}

	integration := query.AIIntegration
	if integration == nil {
		return diag.Errorf("could not find AI integration")
	}

	d.SetId(integration.ID)

	flattened := flattenAIIntegration(integration)
	baseURL := flattened["base_url"]
	delete(flattened, "base_url")

	for key, value := range flattened {
		d.Set(key, value)
	}

	// A provider this version does not model is left without a block rather
	// than failing the read.
	block, ok := aiProviderBlocks[integration.Provider]
	if !ok {
		return nil
	}

	if integration.Provider != aiProviderBedrock {
		d.Set(block, []any{map[string]any{"base_url": baseURL}})
		return nil
	}

	entry := map[string]any{
		"integration_id":   flattened["provider_integration_id"],
		"integration_name": "",
		"region":           "",
		"profiles":         []string{},
	}

	if integration.ProviderIntegrationName != nil {
		entry["integration_name"] = *integration.ProviderIntegrationName
	}

	if integration.ProviderConfig != nil {
		entry["region"] = integration.ProviderConfig.Bedrock.Region
		entry["profiles"] = integration.ProviderConfig.Bedrock.Profiles
	}

	d.Set(block, []any{entry})

	return nil
}
