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

func resourceNamedWebhookSecretHeader() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_named_webhook_secret_header` represents secret key value combination used as a custom header" +
			"when delivering webhook requests. It depends on `spacelift_named_webhook` resource which should exist.",

		CreateContext: resourceNamedWebhookSecretHeaderCreate,
		ReadContext:   resourceNamedWebhookSecretHeaderRead,
		DeleteContext: resourceNamedWebhookSecretHeaderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"webhook_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack on which the environment variable is defined",
				Required:    true,
				ForceNew:    true,
			},
			"key": {
				Type:             schema.TypeString,
				Description:      "key for the header",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ForceNew:         true,
			},
			"value": {
				Type:             schema.TypeString,
				Description:      "value for the header",
				DiffSuppressFunc: suppressValueChange,
				Required:         true,
				Sensitive:        true,
				ForceNew:         true,
			},
		},
	}
}

func resourceNamedWebhookSecretHeaderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		WebhooksIntegration struct {
			ID string `graphql:"id"`
		} `graphql:"namedWebhooksIntegrationSetHeaders(id: $id, input: $input)"`
	}

	webhookID, _ := d.GetOk("webhook_id")
	variables := map[string]interface{}{
		"id": toID(webhookID),
		"input": structs.NamedWebhookHeaderInput{
			Entries: []structs.KeyValuePair{
				{
					Key:   d.Get("key").(string),
					Value: d.Get("value").(string),
				},
			},
		},
	}

	var ret diag.Diagnostics
	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookSecretHeaderSet", &mutation, variables); err != nil {
		ret = diag.Errorf("could not set secret header: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("webhook_id"), d.Get("key")))

	return append(ret, resourceNamedWebhookSecretHeaderRead(ctx, d, meta)...)
}

func resourceNamedWebhookSecretHeaderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	var query struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegration(id: $id)"`
	}

	resourceID, variableKey := idParts[0], idParts[1]

	variables := map[string]interface{}{"id": graphql.ID(resourceID)}
	if err := meta.(*internal.Client).Query(ctx, "GetNamedWebhook", &query, variables); err != nil {
		return diag.Errorf("could not query for named webhook: %v", err)
	}

	if query.Webhook == nil {
		d.SetId("")
		return nil
	}

	wh := query.Webhook

	found := false
	for _, sh := range wh.SecretHeaders {
		if sh == variableKey {
			d.Set("key", sh)
			found = true
			break
		}
	}

	// If we didn't fail adding the key this should never happen.
	if !found {
		d.SetId("")
	}

	return nil
}

func resourceNamedWebhookSecretHeaderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	var mutation struct {
		WebhooksIntegration struct {
			ID string `graphql:"id"`
		} `graphql:"namedWebhooksIntegrationDeleteHeaders(id: $id, headerKeys: $headerKeys)"`
	}

	resourceID, variableKey := idParts[0], idParts[1]
	variables := map[string]interface{}{
		"id":         toID(resourceID),
		"headerKeys": []graphql.String{graphql.String(variableKey)},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "NamedWebhookSecretHeaderDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete named webhook: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")
	return nil
}
