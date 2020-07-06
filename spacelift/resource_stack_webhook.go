package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceStackWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackWebhookCreate,
		Read:   resourceStackWebhookRead,
		Update: resourceStackWebhookUpdate,
		Delete: resourceStackWebhookDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"deleted": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "is deleted",
				Computed:    true,
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks",
				Optional:    true,
				Default:     true,
			},
			"endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "endpoint to send the POST request to",
				Required:    true,
			},
			"secret": &schema.Schema{
				Type:        schema.TypeString,
				Description: "secret used to sign each POST request so you're able to verify that the request comes from us",
				Optional:    true,
				Default:     "",
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which triggers the webhooks",
				Required:    true,
			},
		},
	}
}

func resourceStackWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	secret := d.Get("secret").(string)
	stackID := d.Get("stack_id").(string)

	var mutation struct {
		WebhooksIntegration struct {
			Id      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"webhooksIntegrationCreate(stack: $stack, input: $input)"`
	}

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"input": structs.WebhooksIntegrationInput{
			Enabled:  graphql.Boolean(enabled),
			Endpoint: graphql.String(endpoint),
			Secret:   graphql.String(secret),
		},
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create webhook on the stack")
	}

	if !mutation.WebhooksIntegration.Enabled {
		return errors.New("webhook not activated")
	}

	d.SetId(mutation.WebhooksIntegration.Id)
	d.Set("deleted", false)

	return nil
}

func resourceStackWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{
		"id": toID(stackID),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		return errors.New("stack not found")
	}

	webhookID := d.Id()

	webhookIndex := -1
	for i, webhook := range stack.Integrations.Webhooks {
		if webhook.ID == webhookID {
			webhookIndex = i
			break
		}
	}
	if webhookIndex == -1 {
		d.SetId("")
		return nil
	}

	d.SetId(webhookID)
	d.Set("deleted", stack.Integrations.Webhooks[webhookIndex].Deleted)
	d.Set("enabled", stack.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", stack.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", stack.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}

func resourceStackWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	secret := d.Get("secret").(string)
	stackID := d.Get("stack_id").(string)
	webhookID := d.Id()

	var mutation struct {
		WebhooksIntegration struct {
		} `graphql:"webhooksIntegrationUpdate(stack: $stack, id: $webhook, input: $input)"`
	}

	variables := map[string]interface{}{
		"stack":   toID(stackID),
		"webhook": toID(webhookID),
		"input": structs.WebhooksIntegrationInput{
			Enabled:  graphql.Boolean(enabled),
			Endpoint: graphql.String(endpoint),
			Secret:   graphql.String(secret),
		},
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update webhook")
	}

	return nil
}

func resourceStackWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		WebhooksIntegration struct {
		} `graphql:"webhooksIntegrationDelete(stack: $stack, id: $webhook)"`
	}

	variables := map[string]interface{}{
		"stack":   toID(d.Get("stack_id").(string)),
		"webhook": toID(d.Id()),
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete webhook")
	}

	d.SetId("")
	return nil
}
