package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_webhook` represents a webhook endpoint to which Spacelift " +
			"sends the POST request about run state changes.",

		ReadContext: dataWebhookRead,

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
				Type:         schema.TypeString,
				Description:  "ID of the stack which triggers the webhooks",
				Optional:     true,
				ExactlyOneOf: []string{"module_id", "stack_id"},
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

func dataWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("module_id"); ok {
		return dataModuleWebhookRead(ctx, d, meta)
	}

	return dataStackWebhookRead(ctx, d, meta)
}

func dataModuleWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	moduleID := d.Get("module_id").(string)
	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(moduleID)}

	if err := meta.(*internal.Client).Query(ctx, "ModuleWebhookRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
	}

	module := query.Module
	if module == nil {
		return diag.Errorf("module not found")
	}

	webhookIndex := -1
	for i, webhook := range module.Integrations.Webhooks {
		if webhook.ID == webhookID {
			webhookIndex = i
			break
		}
	}
	if webhookIndex == -1 {
		return diag.Errorf("webhook not found")
	}

	d.SetId(webhookID)
	d.Set("enabled", module.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", module.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", module.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}
func dataStackWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id").(string)
	webhookID := d.Get("webhook_id").(string)
	variables := map[string]interface{}{"id": toID(stackID)}

	if err := meta.(*internal.Client).Query(ctx, "StackWebhookRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		return diag.Errorf("stack not found")
	}

	webhookIndex := -1
	for i, webhook := range stack.Integrations.Webhooks {
		if webhook.ID == webhookID {
			webhookIndex = i
			break
		}
	}
	if webhookIndex == -1 {
		return diag.Errorf("webhook not found")
	}

	d.SetId(webhookID)
	d.Set("enabled", stack.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", stack.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", stack.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}
