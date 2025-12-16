package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceNamedWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_named_webhook` represents a named webhook endpoint used for creating webhooks" +
			"which are referred to in Notification policies to route messages.",

		CreateContext: resourceNamedWebhookCreate,
		ReadContext:   resourceNamedWebhookRead,
		UpdateContext: resourceNamedWebhookUpdate,
		DeleteContext: resourceNamedWebhookDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "enables or disables sending webhooks.",
				Required:    true,
			},
			"endpoint": {
				Type:             schema.TypeString,
				Description:      "endpoint to send the requests to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space_id": {
				Type:             schema.TypeString,
				Description:      "ID of the space the webhook is in",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "the name for the webhook which will also be used to generate the id",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "labels for the webhook to use when referring in policies or filtering them",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"secret": {
				Type:             schema.TypeString,
				Description:      "secret used to sign each request so you're able to verify that the request comes from us. Defaults to an empty value. Note that once it's created, it will be just an empty string in the state due to security reasons.",
				Optional:         true,
				Sensitive:        true,
				DiffSuppressFunc: ignoreOnceCreated,
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

func resourceNamedWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"namedWebhooksIntegrationCreate(input: $input)"`
	}
	input := structs.NamedWebhooksIntegrationInput{
		Enabled:  graphql.Boolean(d.Get("enabled").(bool)),
		Endpoint: graphql.String(d.Get("endpoint").(string)),
		Space:    graphql.ID(d.Get("space_id").(string)),
		Name:     graphql.String(d.Get("name").(string)),
		Secret:   toOptionalString(d.Get("secret").(string)),
		Labels:   []graphql.String{},
	}

	if retryOnFailure, ok := d.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	variables := map[string]interface{}{"input": input}

	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create named webhook: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.WebhooksIntegration.ID)
	return resourceNamedWebhookRead(ctx, d, meta)
}

func resourceNamedWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "GetNamedWebhook", &query, variables); err != nil {
		return diag.Errorf("could not query for named webhook: %v", err)
	}

	if query.Webhook == nil {
		d.SetId("")
		return nil
	}

	wh := query.Webhook
	d.SetId(wh.ID)
	d.Set("name", wh.Name)
	d.Set("endpoint", wh.Endpoint)
	d.Set("enabled", wh.Enabled)
	d.Set("space_id", wh.Space.ID)
	d.Set("secret", "")
	d.Set("retry_on_failure", wh.RetryOnFailure)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range wh.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceNamedWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookID := d.Id()

	enabled := d.Get("enabled").(bool)
	endpoint := d.Get("endpoint").(string)
	spaceID := d.Get("space_id").(string)
	name := d.Get("name").(string)

	labels := []graphql.String{}
	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	var mutation struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegrationUpdate(id: $id, input: $input)"`
	}

	input := structs.NamedWebhooksIntegrationInput{
		Enabled:  graphql.Boolean(enabled),
		Endpoint: graphql.String(endpoint),
		Name:     graphql.String(name),
		Space:    graphql.String(spaceID),
		Labels:   labels,
	}

	if retryOnFailure, ok := d.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	variables := map[string]interface{}{
		"id":    toID(webhookID),
		"input": input,
	}

	var ret diag.Diagnostics
	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update named webhook: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceNamedWebhookRead(ctx, d, meta)...)
}

func resourceNamedWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WebhooksIntegration struct {
			ID string `graphql:"id"`
		} `graphql:"namedWebhooksIntegrationDelete(id: $webhook)"`
	}

	variables := map[string]interface{}{
		"webhook": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete named webhook: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
