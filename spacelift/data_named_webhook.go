package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataNamedWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_named_webhook` represents a named webhook endpoint used for creating webhooks" +
			"which are referred to in Notification policies to route messages.",

		ReadContext: dataNamedWebhookRead,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks.",
				Computed:    true,
			},
			"endpoint": {
				Type:        schema.TypeString,
				Description: "endpoint to send the requests to",
				Computed:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID of the space the webhook is in",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "the name for the webhook which will also be used to generate the id",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "labels for the webhook to use when referring in policies or filtering them",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"secret": {
				Type:        schema.TypeString,
				Description: "secret used to sign each request so you're able to verify that the request comes from us.",
				Computed:    true,
				Sensitive:   true,
			},
			"webhook_id": {
				Type:             schema.TypeString,
				Description:      "ID of the webhook",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
		},
	}
}

func dataNamedWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegration(id: $id)"`
	}

	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(webhookID)}
	if err := meta.(*internal.Client).Query(ctx, "GetNamedWebhook", &query, variables); err != nil {
		return diag.Errorf("could not query for named webhook: %v", err)
	}

	if query.Webhook == nil {
		return diag.Errorf("could not find named webhook")
	}

	wh := query.Webhook
	d.SetId(wh.ID)
	d.Set("name", wh.Name)
	d.Set("endpoint", wh.Endpoint)
	d.Set("secret", wh.Secret)
	d.Set("enabled", wh.Enabled)
	d.Set("space_id", wh.Space.ID)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range wh.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}
