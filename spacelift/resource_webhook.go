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
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_webhook` represents a webhook endpoint to which Spacelift " +
			"sends the POST request about run state changes.",

		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
				ID := d.Id()

				parts := strings.Split(ID, "/")

				if len(parts) != 3 {
					return nil, fmt.Errorf("invalid ID: expected [stack|module]/$projectId/$webhookId, got %q", ID)
				}

				resourceType, resourceID, webhookID := parts[0], parts[1], parts[2]

				switch resourceType {
				case "module":
					d.Set("module_id", resourceID)
				case "stack":
					d.Set("stack_id", resourceID)
				default:
					return nil, fmt.Errorf("invalid resource type %q, only module and stack are supported", resourceType)
				}

				d.SetId(webhookID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks. Defaults to `true`.",
				Optional:    true,
				Default:     true,
			},
			"endpoint": {
				Type:             schema.TypeString,
				Description:      "endpoint to send the POST request to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module which triggers the webhooks",
				Optional:     true,
				ExactlyOneOf: []string{"module_id", "stack_id"},
			},
			"secret": {
				Type:             schema.TypeString,
				Description:      "secret used to sign each POST request so you're able to verify that the request comes from us. Defaults to an empty value. Note that once it's created, it will be just an empty string in the state due to security reasons.",
				Optional:         true,
				Sensitive:        true,
				ForceNew:         true,
				Default:          "",
				DiffSuppressFunc: ignoreOnceCreated,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which triggers the webhooks",
				Optional:    true,
			},
			"retry_on_failure": {
				Type:        schema.TypeBool,
				Description: "whether to retry the webhook in case of failure. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"webhooksIntegrationCreate(stack: $stack, input: $input)"`
	}

	input := structs.WebhooksIntegrationInput{
		Enabled:  graphql.Boolean(d.Get("enabled").(bool)),
		Endpoint: graphql.String(d.Get("endpoint").(string)),
		Secret:   graphql.String(d.Get("secret").(string)),
	}

	if retryOnFailure, ok := d.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	variables := map[string]interface{}{
		"input": input,
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(stackID)
	} else {
		moduleID := d.Get("module_id").(string)
		if err := verifyModule(ctx, moduleID, meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(moduleID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "WebhookCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create webhook: %v", internal.FromSpaceliftError(err))
	}

	if !mutation.WebhooksIntegration.Enabled {
		return diag.Errorf("webhook not activated")
	}

	d.SetId(mutation.WebhooksIntegration.ID)

	return resourceWebhookRead(ctx, d, meta)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleWebhookRead(ctx, d, meta)
	}

	return resourceStackWebhookRead(ctx, d, meta)
}

func resourceModuleWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Get("module_id")),
	}

	if err := meta.(*internal.Client).Query(ctx, "ModuleWebhookRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
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
	d.Set("enabled", module.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", module.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", "")
	d.Set("retry_on_failure", module.Integrations.Webhooks[webhookIndex].RetryOnFailure)

	return nil
}

func resourceStackWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Get("stack_id")),
	}

	if err := meta.(*internal.Client).Query(ctx, "StackWebhookRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
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
	d.Set("enabled", stack.Integrations.Webhooks[webhookIndex].Enabled)
	d.Set("endpoint", stack.Integrations.Webhooks[webhookIndex].Endpoint)
	d.Set("secret", "")
	d.Set("retry_on_failure", stack.Integrations.Webhooks[webhookIndex].RetryOnFailure)

	return nil
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	webhookID := d.Id()

	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"webhooksIntegrationUpdate(stack: $stack, id: $webhook, input: $input)"`
	}

	input := structs.WebhooksIntegrationInput{
		Enabled:  graphql.Boolean(enabled),
		Endpoint: graphql.String(endpoint),
	}

	if retryOnFailure, ok := d.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	variables := map[string]interface{}{
		"webhook": toID(webhookID),
		"input":   input,
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else {
		variables["stack"] = toID(d.Get("module_id"))
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "WebhookUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update webhook: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	} else {
		variables["stack"] = toID(d.Get("module_id"))
	}

	if err := meta.(*internal.Client).Mutate(ctx, "WebhookDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete webhook: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
