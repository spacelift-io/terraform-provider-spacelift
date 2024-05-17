package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceAuditTrailWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_audit_trail_webhook` represents a webhook endpoint to which Spacelift " +
			"sends POST requests about audit events.",
		CreateContext: resourceAuditTrailWebhookCreate,
		ReadContext:   resourceAuditTrailWebhookRead,
		UpdateContext: resourceAuditTrailWebhookUpdate,
		DeleteContext: resourceAuditTrailWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
				Description: "`enabled` determines whether the webhook is enabled. If it is not, " +
					"Spacelift will not send any requests to the endpoint.",
			},
			"endpoint": {
				Type:     schema.TypeString,
				Required: true,
				Description: "`endpoint` is the URL to which Spacelift will send POST requests " +
					"about audit events.",
			},
			"include_runs": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "`include_runs` determines whether the webhook should include " +
					"information about the run that triggered the event.",
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "`secret` is a secret that Spacelift will send with the request.",
			},
			"custom_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "`custom_headers` is a Map of key values strings, that will be passed as headers with audit trail call.",
			},
		},
	}
}

func resourceAuditTrailWebhookCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhook `graphql:"auditTrailSetWebhook(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": structs.AuditTrailWebhookInput{
			Enabled:       toBool(data.Get("enabled")),
			Endpoint:      toString(data.Get("endpoint")),
			IncludeRuns:   toBool(data.Get("include_runs")),
			Secret:        toString(data.Get("secret")),
			CustomHeaders: toOptionalStringMap(data.Get("custom_headers")),
		},
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	data.SetId(time.Now().String())

	return resourceAuditTrailWebhookRead(ctx, data, i)
}

func resourceAuditTrailWebhookRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var query struct {
		AuditTrailWebhook *structs.AuditTrailWebhook `graphql:"auditTrailWebhook"`
	}
	if err := i.(*internal.Client).Query(ctx, "AuditTrailWebhookRead", &query, nil); err != nil {
		return diag.Errorf("could not query for audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	if query.AuditTrailWebhook == nil {
		data.SetId("")
		return nil
	}

	data.Set("enabled", query.AuditTrailWebhook.Enabled)
	data.Set("endpoint", query.AuditTrailWebhook.Endpoint)
	data.Set("include_runs", query.AuditTrailWebhook.IncludeRuns)
	data.Set("secret", query.AuditTrailWebhook.Secret)
	data.Set("custom_headers", query.AuditTrailWebhook.CustomHeaders.ToStdMap())

	return nil
}

func resourceAuditTrailWebhookUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhook `graphql:"auditTrailSetWebhook(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": structs.AuditTrailWebhookInput{
			Enabled:       toBool(data.Get("enabled")),
			Endpoint:      toString(data.Get("endpoint")),
			IncludeRuns:   toBool(data.Get("include_runs")),
			Secret:        toString(data.Get("secret")),
			CustomHeaders: toOptionalStringMap(data.Get("custom_headers")),
		},
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	return resourceAuditTrailWebhookRead(ctx, data, i)
}

func resourceAuditTrailWebhookDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhook `graphql:"auditTrailDeleteWebhook"`
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookDelete", &mutation, nil); err != nil {
		return diag.Errorf("could not delete audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	data.SetId("")

	return nil
}
