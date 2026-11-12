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

func resourceSlackChannelIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_slack_channel_integration` represents a Slack channel integration " +
			"that grants space-level permissions to members of a Slack channel.\n\n" +
			"The Slack workspace must be connected via OAuth in the Spacelift UI " +
			"before using this resource.",
		CreateContext: resourceSlackChannelIntegrationCreate,
		ReadContext:   resourceSlackChannelIntegrationRead,
		UpdateContext: resourceSlackChannelIntegrationUpdate,
		DeleteContext: resourceSlackChannelIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"integration_name": {
				Type:             schema.TypeString,
				Description:      "Human-readable name for the integration",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"slack_channel_id": {
				Type:             schema.TypeString,
				Description:      "ID of the Slack channel to grant permissions to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"access_rule": {
				Type:        schema.TypeSet,
				Description: "List of space access rules for the Slack channel integration. At least one rule is required.",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:             schema.TypeString,
							Description:      "ID (slug) of the space the integration has access to",
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
		},
	}
}

func resourceSlackChannelIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		Integration *structs.ManagedUserGroupIntegration `graphql:"managedUserGroupIntegrationCreate(input: $input)"`
	}
	variables := map[string]any{
		"input": structs.ManagedUserGroupIntegrationCreateInput{
			IntegrationType: graphql.String("SLACK"),
			IntegrationName: toString(d.Get("integration_name")),
			SlackChannelID:  toString(d.Get("slack_channel_id")),
			AccessRules:     getSlackChannelIntegrationAccessRules(d),
		},
	}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create Slack channel integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.Integration.ID)

	return resourceSlackChannelIntegrationRead(ctx, d, meta)
}

func resourceSlackChannelIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		Integration *structs.ManagedUserGroupIntegration `graphql:"managedUserGroupIntegration(id: $id)"`
	}
	variables := map[string]any{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "ManagedUserGroupIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for Slack channel integration: %v", err)
	}

	integration := query.Integration
	if integration == nil {
		d.SetId("")
		return nil
	}

	d.Set("integration_name", integration.IntegrationName)
	d.Set("slack_channel_id", integration.SlackChannelID)

	var accessList []any
	for _, a := range integration.AccessRules {
		accessList = append(accessList, map[string]any{
			"space_id": a.Space,
			"role":     a.SpaceAccessLevel,
		})
	}
	d.Set("access_rule", accessList)

	return nil
}

func resourceSlackChannelIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var ret diag.Diagnostics

	var mutation struct {
		Integration *structs.ManagedUserGroupIntegration `graphql:"managedUserGroupIntegrationUpdate(input: $input)"`
	}
	variables := map[string]any{
		"input": structs.ManagedUserGroupIntegrationUpdateInput{
			ID:              toID(d.Id()),
			IntegrationName: toString(d.Get("integration_name")),
			SlackChannelID:  toString(d.Get("slack_channel_id")),
			AccessRules:     getSlackChannelIntegrationAccessRules(d),
		},
	}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupIntegrationUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update Slack channel integration: %v", internal.FromSpaceliftError(err))...)
	}

	ret = append(ret, resourceSlackChannelIntegrationRead(ctx, d, meta)...)

	return ret
}

func resourceSlackChannelIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		Integration *structs.ManagedUserGroupIntegration `graphql:"managedUserGroupIntegrationDelete(id: $id)"`
	}
	variables := map[string]any{"id": toID(d.Id())}
	if err := meta.(*internal.Client).Mutate(ctx, "ManagedUserGroupIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete Slack channel integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func getSlackChannelIntegrationAccessRules(d *schema.ResourceData) []structs.SpaceAccessRuleInput {
	accessRules := make([]structs.SpaceAccessRuleInput, 0)

	if rules, ok := d.Get("access_rule").(*schema.Set); ok {
		for _, a := range rules.List() {
			access := a.(map[string]any)
			accessRules = append(accessRules, structs.SpaceAccessRuleInput{
				Space:            toID(access["space_id"]),
				SpaceAccessLevel: structs.SpaceAccessLevel(access["role"].(string)),
			})
		}
	}

	return accessRules
}
