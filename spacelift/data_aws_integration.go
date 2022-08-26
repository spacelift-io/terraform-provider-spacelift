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

func dataAWSIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_integration` represents an integration with an AWS " +
			"account. This integration is account-level and needs to be explicitly " +
			"attached to individual stacks in order to take effect." +
			"\n\n" +
			"Note: when assuming credentials for **shared workers**, Spacelift will use `$accountName-$integrationID@$stackID-suffix` " +
			"or `$accountName-$integrationID@$moduleID-$suffix` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) " +
			"and `$runID@$stackID@$accountName` truncated to 64 characters as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole)," +
			"$suffix will be `read` or `write`.",

		ReadContext: dataAWSIntegrationRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:             schema.TypeString,
				Description:      "Immutable ID of the integration. Either `integration_id` or `name` must be specified.",
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ExactlyOneOf:     []string{"integration_id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the AWS integration. Either `integration_id` or `name` must be specified.",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"integration_id", "name"},
			},
			"role_arn": {
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Computed:    true,
			},
			"duration_seconds": {
				Type:        schema.TypeInt,
				Description: "Duration in seconds for which the assumed role credentials should be valid",
				Computed:    true,
			},
			"generate_credentials_in_worker": {
				Type:        schema.TypeBool,
				Description: "Generate AWS credentials in the private worker",
				Computed:    true,
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "Custom external ID (works only for private workers).",
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the integration is in",
				Computed:    true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataAWSIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	integrationID := d.Get("integration_id").(string)
	name := d.Get("name").(string)
	client := meta.(*internal.Client)

	var diagnostics diag.Diagnostics
	var integration *structs.AWSIntegration
	if integrationID != "" {
		integration, diagnostics = findAWSIntegrationByID(ctx, integrationID, client)
	} else {
		integration, diagnostics = findAWSIntegrationByName(ctx, name, client)
	}

	if diagnostics != nil {
		return diagnostics
	}

	if integration == nil {
		idOrName := integrationID
		if idOrName == "" {
			idOrName = name
		}

		return diag.Errorf("AWS integration not found: %s", idOrName)
	}

	d.SetId(integration.ID)
	integration.PopulateResourceData(d)

	return nil
}

func findAWSIntegrationByID(ctx context.Context, integrationID string, client *internal.Client) (*structs.AWSIntegration, diag.Diagnostics) {
	var query struct {
		AWSIntegration *structs.AWSIntegration `graphql:"awsIntegration(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.ID(integrationID)}
	if err := client.Query(ctx, "AWSIntegrationRead", &query, variables); err != nil {
		return nil, diag.Errorf("could not query for the aws integration: %v", err)
	}

	return query.AWSIntegration, nil
}

func findAWSIntegrationByName(ctx context.Context, name string, client *internal.Client) (*structs.AWSIntegration, diag.Diagnostics) {
	var query struct {
		AWSIntegration *structs.AWSIntegration `graphql:"awsIntegrationByName(name: $name)"`
	}
	variables := map[string]interface{}{"name": graphql.String(name)}
	if err := client.Query(ctx, "AWSIntegrationByNameRead", &query, variables); err != nil {
		return nil, diag.Errorf("could not query for the aws integration: %v", err)
	}

	return query.AWSIntegration, nil
}
