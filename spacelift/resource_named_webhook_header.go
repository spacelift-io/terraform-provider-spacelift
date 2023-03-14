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

func resourceNamedWebhookHeader() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_named_webhook_header` represents a single header sent " +
			"with the webhook payload. These are normally used to authenticate " +
			"or authorize the webhook request.",
		CreateContext: resourceNamedWebhookHeaderCreate,
		ReadContext:   resourceNamedWebhookHeaderRead,
		DeleteContext: resourceNamedWebhookHeaderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"named_webhook_id": {
				Type:        schema.TypeString,
				Description: "ID of the named webhook to which this header belongs",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the header",
				Required:    true,
				ForceNew:    true,
			},
			"value": {
				Type:        schema.TypeString,
				Description: "Value of the header",
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diffs if the old value is empty, meaning that
					// the header was actually imported, not created using Terraform.
					return old == ""
				},
			},
		},
	}
}

func resourceNamedWebhookHeaderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WebhooksIntegration *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegrationSetHeaders(id: $id, input: $input)"`
	}

	webhookID := d.Get("named_webhook_id").(string)
	headerName := d.Get("name").(string)

	input := structs.NamedWebhookHeaderInput{
		Entries: []structs.NamedWebhookHeaderInputEntry{{
			Key:   headerName,
			Value: d.Get("value").(string),
		}},
	}

	variables := map[string]interface{}{
		"id":    graphql.ID(webhookID),
		"input": input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookHeaderCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create named webhook: %v", internal.FromSpaceliftError(err))
	}

	headerID := fmt.Sprintf("%s/%s", webhookID, d.Get("name").(string))
	d.SetId(headerID)

	return nil
}

func resourceNamedWebhookHeaderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("invalid ID format: %s", d.Id())
	}

	var query struct {
		WebhooksIntegration *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegration(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(idParts[0]),
	}

	if err := meta.(*internal.Client).Query(ctx, "NamedWebhookHeaderRead", &query, variables); err != nil {
		return diag.Errorf("could not query for named webhook: %v", internal.FromSpaceliftError(err))
	}

	if query.WebhooksIntegration == nil {
		// The integration was deleted, so the header is gone as well.
		d.SetId("")
		return nil
	}

	var found bool
	for _, header := range query.WebhooksIntegration.SecretHeaders {
		if header == idParts[1] {
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}

func resourceNamedWebhookHeaderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("invalid ID format: %s", d.Id())
	}

	if readOp := resourceNamedWebhookHeaderRead(ctx, d, meta); readOp.HasError() {
		return readOp
	}

	var mutation struct {
		WebhooksIntegration *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegrationRemoveHeaders(id: $id, headerKeys: $headerKeys)"`
	}

	variables := map[string]interface{}{
		"id":         graphql.ID(idParts[0]),
		"headerKeys": []string{idParts[1]},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookHeaderDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete named webhook header: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
