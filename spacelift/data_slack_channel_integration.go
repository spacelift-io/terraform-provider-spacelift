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

func dataSlackChannelIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_slack_channel_integration` represents a data source for retrieving " +
			"information about an existing Slack channel integration in Spacelift.",

		ReadContext: dataSlackChannelIntegrationRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "ID of the Slack channel integration",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"integration_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-readable name for the integration",
			},
			"slack_channel_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the Slack channel",
			},
			"access_rule": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of space access rules for the Slack channel integration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"space_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID (slug) of the space the integration has access to",
						},
						"role": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of access to the space. Possible values are: READ, WRITE, ADMIN",
						},
					},
				},
				Set: userPolicyHash,
			},
		},
	}
}

func dataSlackChannelIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	integrationID := d.Get("integration_id").(string)

	var query struct {
		Integration *structs.ManagedUserGroupIntegration `graphql:"managedUserGroupIntegration(id: $id)"`
	}
	variables := map[string]any{"id": graphql.ID(integrationID)}
	if err := meta.(*internal.Client).Query(ctx, "ManagedUserGroupIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for Slack channel integration: %v", err)
	}

	integration := query.Integration
	if integration == nil {
		return diag.Errorf("could not find Slack channel integration with ID %q", integrationID)
	}

	d.SetId(integration.ID)
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
