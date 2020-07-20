package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataStackWebhook() *schema.Resource {
	return &schema.Resource{
		Read: dataStackWebhookRead,
		Schema: map[string]*schema.Schema{
			"deleted": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "is deleted",
				Computed:    true,
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks",
				Computed:    true,
			},
			"endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "endpoint to send the POST request to",
				Computed:    true,
			},
			"secret": &schema.Schema{
				Type:        schema.TypeString,
				Description: "secret used to sign each POST request so you're able to verify that the request comes from us",
				Computed:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which triggers the webhooks",
				Required:    true,
			},
			"webhook_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the webhook",
				Required:    true,
			},
		},
	}
}

func dataStackWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id").(string)
	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(stackID)}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		return errors.New("stack not found")
	}

	webhookIndex := -1
	for i, webhook := range stack.Integrations.Webhooks {
		if webhook.ID == webhookID {
			webhookIndex = i
			break
		}
	}
	if webhookIndex == -1 {
		return errors.New("webhook not found")
	}

	d.SetId(webhookID)
	d.Set("deleted", stack.Integrations.Webhooks[webhookIndex].Deleted)
	d.Set("enabled", stack.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", stack.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", stack.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}
