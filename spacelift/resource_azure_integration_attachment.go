package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceAzureIntegrationAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_azure_integration_attachment` represents the attachment between " +
			"a reusable Azure integration and a single stack or module.",

		CreateContext: resourceAzureIntegrationAttachmentCreate,
		ReadContext:   resourceAzureIntegrationAttachmentRead,
		UpdateContext: resourceAzureIntegrationAttachmentUpdate,
		DeleteContext: resourceAzureIntegrationAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Description: "ID of the integration to attach",
				Required:    true,
				ForceNew:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module to attach the integration to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"stack_id": {
				Type:         schema.TypeString,
				Description:  "ID of the stack to attach the integration to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"read": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this attachment is used for read operations",
				Optional:    true,
				Default:     true,
			},
			"subscription_id": {
				Type: schema.TypeString,
				Description: "" +
					"Contains the Azure subscription ID to use with this Stack. " +
					" Overrides the default subscription ID set at the integration " +
					"level.",
				Optional: true,
			},
			"write": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this attachment is used for write operations",
				Default:     true,
				Optional:    true,
			},
			"attachment_id": {
				Type:        schema.TypeString,
				Description: "Internal ID of the attachment entity",
				Computed:    true,
			},
		},
	}
}

func resourceAzureIntegrationAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AzureIntegrationAttach structs.AzureIntegrationAttachment `graphql:"azureIntegrationAttach(id: $id, stack: $projectId, read: $read, write: $write, subscriptionId: $subscriptionId)"`
	}

	projectID := projectID(d)

	variables := map[string]interface{}{
		"id":             toID(d.Get("integration_id")),
		"projectId":      projectID,
		"read":           graphql.Boolean(d.Get("read").(bool)),
		"write":          graphql.Boolean(d.Get("write").(bool)),
		"subscriptionId": (*graphql.String)(nil),
	}

	if subscriptionID, ok := d.GetOk("subscription_id"); ok {
		variables["subscriptionId"] = toOptionalString(subscriptionID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationAttachmentCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not attach the Azure integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("integration_id"), projectID))

	return resourceAzureIntegrationAttachmentRead(ctx, d, meta)
}

func resourceAzureIntegrationAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureIntegration *struct {
			Attachment *structs.AzureIntegrationAttachment `graphql:"attachedStack(id: $projectId)"`
		} `graphql:"azureIntegration(id: $integrationId)"`
	}

	idComponents := strings.SplitN(d.Id(), "/", 2)
	if len(idComponents) != 2 {
		return diag.Errorf("invalid ID: %s", d.Id())
	}

	variables := map[string]interface{}{
		"integration_id": toID(idComponents[0]),
		"project_id":     toID(idComponents[1]),
	}

	if err := meta.(*internal.Client).Query(ctx, "AzureIntegrationAttachmentRead", &query, variables); err != nil {
		return diag.FromErr(err)
	}

	if query.AzureIntegration == nil || query.AzureIntegration.Attachment == nil {
		d.SetId("")
		return nil
	}

	query.AzureIntegration.Attachment.PopulateResourceData(d)

	return nil
}

func resourceAzureIntegrationAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AzureIntegrationAttachmentUpdate structs.AzureIntegrationAttachment `graphql:"azureIntegrationAttachmentUpdate(id: $id, read: $read, write: $write, subscriptionId: $subscriptionId)"`
	}

	variables := map[string]interface{}{
		"id":             toID(d.Get("attachment_id")),
		"read":           graphql.Boolean(d.Get("read").(bool)),
		"write":          graphql.Boolean(d.Get("write").(bool)),
		"subscriptionId": (*graphql.String)(nil),
	}

	if subscriptionID, ok := d.GetOk("subscription_id"); ok {
		variables["subscriptionId"] = toOptionalString(subscriptionID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationAttachmentUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update the Azure integration attachment: %v", internal.FromSpaceliftError(err))
	}

	return resourceAzureIntegrationAttachmentRead(ctx, d, meta)
}

func resourceAzureIntegrationAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AzureIntegrationAttachmentDelete struct{} `graphql:"azureIntegrationDetach(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("attachment_id"))}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationAttachmentDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not detach the Azure integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func projectID(d *schema.ResourceData) graphql.ID {
	if moduleID, ok := d.GetOk("module_id"); ok {
		return toID(moduleID)
	}

	return toID(d.Get("stack_id"))
}
