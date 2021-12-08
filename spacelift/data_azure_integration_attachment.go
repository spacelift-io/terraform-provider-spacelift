package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAzureIntegrationAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_azure_integration_attachment` represents the attachment between " +
			"a reusable Azure integration and a single stack or module.",

		ReadContext: dataAzureIntegrationAttachmentRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Description: "ID of the integration to attach",
				Required:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module to attach the integration to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"stack_id": {
				Type:         schema.TypeString,
				Description:  "ID of the stack to attach the integration to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"read": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this attachment is used for read operations",
				Computed:    true,
			},
			"subscription_id": {
				Type: schema.TypeString,
				Description: "" +
					"Contains the Azure subscription ID to use with this Stack. " +
					" Overrides the default subscription ID set at the integration " +
					"level.",
				Computed: true,
			},
			"write": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this attachment is used for write operations",
				Computed:    true,
			},
			"attachment_id": {
				Type:        schema.TypeString,
				Description: "Internal ID of the attachment entity",
				Computed:    true,
			},
		},
	}
}

func dataAzureIntegrationAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureIntegration *struct {
			Attachment *structs.AzureIntegrationAttachment `graphql:"attachedStack(id: $projectId)"`
		} `graphql:"azureIntegration(id: $integrationId)"`
	}

	integrationID := toID(d.Get("integration_id"))
	projectID := toID(projectID(d))

	variables := map[string]interface{}{
		"integration_id": integrationID,
		"project_id":     projectID,
	}

	if err := meta.(*internal.Client).Query(ctx, "AzureIntegrationAttachmentRead", &query, variables); err != nil {
		return diag.FromErr(err)
	}

	if query.AzureIntegration == nil || query.AzureIntegration.Attachment == nil {
		return diag.Errorf("Azure integration attachment not found")
	}

	query.AzureIntegration.Attachment.PopulateResourceData(d)
	d.SetId(fmt.Sprintf("%s/%s", integrationID, projectID))

	return nil
}
