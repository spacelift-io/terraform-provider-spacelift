package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataAWSIntegrationAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_integration_attachment` represents the attachment between " +
			"a reusable AWS integration and a single stack or module.",

		ReadContext: dataAWSIntegrationAttachmentRead,

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

func dataAWSIntegrationAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AWSIntegration *struct {
			Attachment *structs.AWSIntegrationAttachment `graphql:"attachedStack(id: $projectId)"`
		} `graphql:"awsIntegration(id: $integrationId)"`
	}

	integrationID := toID(d.Get("integration_id"))
	projectID := toID(projectID(d))

	variables := map[string]interface{}{
		"integrationId": integrationID,
		"projectId":     projectID,
	}

	if err := meta.(*internal.Client).Query(ctx, "AWSIntegrationAttachmentRead", &query, variables); err != nil {
		return diag.FromErr(err)
	}

	if query.AWSIntegration == nil || query.AWSIntegration.Attachment == nil {
		return diag.Errorf("AWS integration attachment not found")
	}

	query.AWSIntegration.Attachment.PopulateResourceData(d)
	d.SetId(fmt.Sprintf("%s/%s", integrationID, projectID))

	return nil
}
