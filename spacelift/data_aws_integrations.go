package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataAWSIntegrations() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_aws_integration` represents a list of all the AWS integrations in the Spacelift account",
		ReadContext: dataAWSIntegrationsRead,
		Schema: map[string]*schema.Schema{
			"contexts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_id": {
							Type:             schema.TypeString,
							Description:      "Immutable ID of the integration.",
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the AWS integration.",
							Optional:    true,
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
				},
			},
		},
	}
}

func dataAWSIntegrationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Integrations []*structs.AWSIntegration `graphql:"awsIntegrations()"`
	}
	variables := map[string]interface{}{}

	if err := meta.(*internal.Client).Query(ctx, "AwsIntegrationsRead", &query, variables); err != nil {
		return diag.Errorf("could not query for AWS integrations: %v", err)
	}

	d.SetId("spacelift_aws_integrations")

	integrations := query.Integrations
	if integrations == nil {
		d.Set("integrations", nil)
		return nil
	}

	mapped := flattenDataIntegrationsList(integrations)
	if err := d.Set("integrations", mapped); err != nil {
		d.SetId("")
		return diag.Errorf("could not set contexts: %v", err)
	}

	return nil
}

func flattenDataIntegrationsList(integrations []*structs.AWSIntegration) []map[string]interface{} {
	wps := make([]map[string]interface{}, len(integrations))

	for index, integration := range integrations {
		wps[index] = integration.ToMap()
	}

	return wps
}
