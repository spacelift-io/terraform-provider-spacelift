package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebhookCreate,
		Read:   resourceWebhookRead,
		Update: resourceWebhookUpdate,
		Delete: resourceWebhookDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"deleted": {
				Type:        schema.TypeBool,
				Description: "is deleted",
				Computed:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks",
				Optional:    true,
				Default:     true,
			},
			"endpoint": {
				Type:        schema.TypeString,
				Description: "endpoint to send the POST request to",
				Required:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module which triggers the webhooks",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
			},
			"secret": {
				Type:        schema.TypeString,
				Description: "secret used to sign each POST request so you're able to verify that the request comes from us",
				Optional:    true,
				Default:     "",
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which triggers the webhooks",
				Optional:    true,
			},
		},
	}
}

func resourceWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	secret := d.Get("secret").(string)

	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"webhooksIntegrationCreate(stack: $stack, input: $input)"`
	}

	variables := map[string]interface{}{
		"input": structs.WebhooksIntegrationInput{
			Enabled:  graphql.Boolean(enabled),
			Endpoint: graphql.String(endpoint),
			Secret:   graphql.String(secret),
		},
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["stack"] = toID(moduleID)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create webhook")
	}

	if !mutation.WebhooksIntegration.Enabled {
		return errors.New("webhook not activated")
	}

	d.SetId(mutation.WebhooksIntegration.ID)
	d.Set("deleted", false)

	return nil
}

func resourceWebhookRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleWebhookRead(d, meta)
	}

	if _, ok := d.GetOk("stack_id"); ok {
		return resourceStackWebhookRead(d, meta)
	}

	return errors.New("either module_id or stack_id must be provided")
}

func resourceModuleWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Get("module_id")),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	module := query.Module
	if module == nil {
		d.SetId("")
		return nil
	}

	webhookID := d.Id()

	webhookIndex := -1
	for i, webhook := range module.Integrations.Webhooks {
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
	d.Set("deleted", module.Integrations.Webhooks[webhookIndex].Deleted)
	d.Set("enabled", module.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", module.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", module.Integrations.Webhooks[webhookIndex].Secret)

	return nil
}

func resourceStackWebhookRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Get("stack_id")),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		d.SetId("")
		return nil
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

func resourceWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	secret := d.Get("secret").(string)
	webhookID := d.Id()

	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"webhooksIntegrationUpdate(stack: $stack, id: $webhook, input: $input)"`
	}

	variables := map[string]interface{}{
		"webhook": toID(webhookID),
		"input": structs.WebhooksIntegrationInput{
			Enabled:  graphql.Boolean(enabled),
			Endpoint: graphql.String(endpoint),
			Secret:   graphql.String(secret),
		},
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["stack"] = toID(moduleID)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update webhook")
	}

	return nil
}

func resourceWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		WebhooksIntegration struct {
			ID string `graphql:"id"`
		} `graphql:"webhooksIntegrationDelete(stack: $stack, id: $webhook)"`
	}

	variables := map[string]interface{}{
		"webhook": toID(d.Id()),
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["stack"] = toID(moduleID)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete webhook")
	}

	d.SetId("")
	return nil
}
