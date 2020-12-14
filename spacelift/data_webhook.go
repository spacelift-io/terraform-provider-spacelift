package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataWebhook() *schema.Resource {
	return &schema.Resource{
		Read: dataWebhookRead,
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks",
				Computed:    true,
			},
			"endpoint": {
				Type:        schema.TypeString,
				Description: "endpoint to send the POST request to",
				Computed:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the stack which triggers the webhooks",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
			},
			"secret": {
				Type:        schema.TypeString,
				Description: "secret used to sign each POST request so you're able to verify that the request comes from us",
				Computed:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which triggers the webhooks",
				Optional:    true,
			},
			"webhook_id": {
				Type:        schema.TypeString,
				Description: "ID of the webhook",
				Required:    true,
			},
		},
	}
}

func dataWebhookRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("module_id"); ok {
		return dataModuleWebhookRead(d, meta)
	}

	if _, ok := d.GetOk("stack_id"); ok {
		return dataStackWebhookRead(d, meta)
	}

	return errors.New("either module_id or stack_id must be provided")
}

func dataModuleWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id").(string)
	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(moduleID)}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	module := query.Module
	if module == nil {
		return errors.New("module not found")
	}

	webhookIndex := -1
	for i, webhook := range module.Integrations.Webhooks {
		if webhook.ID == webhookID {
			webhookIndex = i
			break
		}
	}
	if webhookIndex == -1 {
		return errors.New("webhook not found")
	}

	d.SetId(webhookID)
	d.Set("enabled", module.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", module.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", module.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}
func dataStackWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id").(string)
	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(stackID)}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
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
	d.Set("enabled", stack.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", stack.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", stack.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}
