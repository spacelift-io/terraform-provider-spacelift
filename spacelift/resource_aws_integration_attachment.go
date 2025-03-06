package spacelift

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceAWSIntegrationAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_integration_attachment` represents the attachment between " +
			"a reusable AWS integration and a single stack or module.",

		CreateContext: resourceAWSIntegrationAttachmentCreate,
		ReadContext:   resourceAWSIntegrationAttachmentRead,
		UpdateContext: resourceAWSIntegrationAttachmentUpdate,
		DeleteContext: resourceAWSIntegrationAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:             schema.TypeString,
				Description:      "ID of the integration to attach",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
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
				Description: "Indicates whether this attachment is used for read operations. Defaults to `true`.",
				Optional:    true,
				Default:     true,
			},
			"write": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this attachment is used for write operations. Defaults to `true`.",
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

func resourceAWSIntegrationAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	projectID := projectID(d)

	var err error
	for i := 0; i < 5; i++ {
		err = resourceAWSIntegrationAttachmentAttach(ctx, meta.(*internal.Client), projectID, d)
		if err == nil || !strings.Contains(err.Error(), "you need to configure trust relationship") || i == 4 {
			break
		}

		// Yay for eventual consistency.
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return diag.Errorf("could not attach the aws integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("integration_id"), projectID))

	return resourceAWSIntegrationAttachmentRead(ctx, d, meta)
}

func resourceAWSIntegrationAttachmentAttach(ctx context.Context, client *internal.Client, projectID graphql.ID, d *schema.ResourceData) error {
	var mutation struct {
		AWSIntegrationAttach structs.AWSIntegrationAttachment `graphql:"awsIntegrationAttach(id: $id, stack: $projectId, read: $read, write: $write)"`
	}

	variables := map[string]interface{}{
		"id":        toID(d.Get("integration_id")),
		"projectId": projectID,
		"read":      graphql.Boolean(d.Get("read").(bool)),
		"write":     graphql.Boolean(d.Get("write").(bool)),
	}

	if err := client.Mutate(ctx, "awsIntegrationAttach", &mutation, variables); err != nil {
		return err
	}

	return nil
}

func resourceAWSIntegrationAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AWSIntegration *struct {
			Attachment *structs.AWSIntegrationAttachment `graphql:"attachedStack(id: $projectId)"`
		} `graphql:"awsIntegration(id: $integrationId)"`
	}

	idComponents := strings.SplitN(d.Id(), "/", 2)
	if len(idComponents) != 2 {
		return diag.Errorf("invalid ID: %s", d.Id())
	}
	integrationID, projectID := idComponents[0], idComponents[1]

	variables := map[string]interface{}{
		"integrationId": toID(integrationID),
		"projectId":     toID(projectID),
	}

	if err := meta.(*internal.Client).Query(ctx, "awsIntegrationAttachmentRead", &query, variables); err != nil {
		return diag.FromErr(err)
	}

	if query.AWSIntegration == nil || query.AWSIntegration.Attachment == nil {
		d.SetId("")
		return nil
	}

	query.AWSIntegration.Attachment.PopulateResourceData(d)

	// This is to allow importing.
	d.Set("integration_id", integrationID)

	return nil
}

func resourceAWSIntegrationAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AWSIntegrationAttachmentUpdate structs.AWSIntegrationAttachment `graphql:"awsIntegrationAttachmentUpdate(id: $id, read: $read, write: $write)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Get("attachment_id")),
		"read":  graphql.Boolean(d.Get("read").(bool)),
		"write": graphql.Boolean(d.Get("write").(bool)),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "awsIntegrationAttachmentUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update the aws integration attachment: %v", internal.FromSpaceliftError(err))
	}

	return resourceAWSIntegrationAttachmentRead(ctx, d, meta)
}

func resourceAWSIntegrationAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AWSIntegrationAttachmentDelete *structs.AWSIntegrationAttachment `graphql:"awsIntegrationDetach(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("attachment_id"))}

	if err := meta.(*internal.Client).Mutate(ctx, "awsIntegrationAttachmentDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not detach the aws integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
