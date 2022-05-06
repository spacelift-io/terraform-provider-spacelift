package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAWSIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_integration` represents an integration with an AWS " +
			"account. This integration is account-level and needs to be explicitly " +
			"attached to individual stacks in order to take effect." +
			"\n\n" +
			"Note: when assuming credentials for **shared worker**, Spacelift will use `$accountName-$integrationID@$stackID-suffix` " +
			"or `$accountName-$integrationID@$moduleID-suffix` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) " +
			"and `$runID@$stackID@$accountName` truncated to 64 characters as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole)," +
			"Suffix will be read or write.",

		ReadContext: dataAWSIntegrationRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Description: "immutable ID of the integration",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the AWS integration",
				Computed:    true,
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
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"legacy": {
				Type:        schema.TypeBool,
				Description: "Indicates if the integration was created via the legacy AWS stack integration functionality",
				Computed:    true,
			},
		},
	}
}

func dataAWSIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AWSIntegration *structs.AWSIntegration `graphql:"awsIntegration(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.ID(d.Get("integration_id").(string))}
	if err := meta.(*internal.Client).Query(ctx, "AWSIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the aws integration: %v", err)
	}

	integration := query.AWSIntegration
	if integration == nil {
		return diag.Errorf("AWS integration not found: %s", d.Id())
	}

	d.SetId(integration.ID)
	integration.PopulateResourceData(d)

	return nil
}
