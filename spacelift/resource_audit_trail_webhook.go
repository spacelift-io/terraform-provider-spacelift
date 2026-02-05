package spacelift

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
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
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				ForceNew:         true,
				Description:      "`secret` is a secret that Spacelift will send with the request. Note that once it's created, it will be just an empty string in the state due to security reasons.",
				DiffSuppressFunc: ignoreOnceCreated,
				ConflictsWith:    []string{"secret_wo", "secret_wo_version"},
			},
			"secret_wo": {
				Type:          schema.TypeString,
				Description:   "Value of the environment variable. The secret is not stored in the state. Modify secret_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
				Sensitive:     true,
				Optional:      true,
				WriteOnly:     true,
				ConflictsWith: []string{"secret"},
				RequiredWith:  []string{"secret_wo_version"},
			},
			"secret_wo_version": {
				Type:          schema.TypeString,
				Description:   "Used together with secret_wo to trigger an update to the secret. Increment this value when an update to secret_wo is required. This field requires Terraform/OpenTofu 1.11+.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"secret"},
				RequiredWith:  []string{"secret_wo_version"},
			},
			"custom_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "`custom_headers` is a Map of key-value strings, that will be passed as headers with audit trail requests.",
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

func resourceAuditTrailWebhookCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var secret string
	if v, ok := data.GetOk("secret"); ok {
		secret = v.(string)
	}

	if _, ok := data.GetOk("secret_wo_version"); ok {
		p := cty.GetAttrPath("secret_wo")
		woVal, d := data.GetRawConfigAt(p)
		if d.HasError() {
			return diag.FromErr(fmt.Errorf("could not get write-only secret: %v", d))
		}

		if !woVal.IsNull() {
			secret = woVal.AsString()
		}
	}

	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhookRead `graphql:"auditTrailSetWebhook(input: $input)"`
	}

	input := structs.AuditTrailWebhookInput{
		Enabled:       toBool(data.Get("enabled")),
		Endpoint:      toString(data.Get("endpoint")),
		IncludeRuns:   toBool(data.Get("include_runs")),
		Secret:        toString(secret),
		CustomHeaders: toOptionalStringMap(data.Get("custom_headers")),
	}

	if retryOnFailure, ok := data.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	variables := map[string]interface{}{
		"input": input,
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	data.SetId(time.Now().String())

	return resourceAuditTrailWebhookRead(ctx, data, i)
}

func resourceAuditTrailWebhookRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var query struct {
		AuditTrailWebhook *structs.AuditTrailWebhookRead `graphql:"auditTrailWebhook"`
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
	data.Set("secret", "")
	data.Set("retry_on_failure", query.AuditTrailWebhook.RetryOnFailure)

	return nil
}

func resourceAuditTrailWebhookUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhookRead `graphql:"auditTrailSetWebhook(input: $input)"`
	}
	input := structs.AuditTrailWebhookInput{
		Enabled:       toBool(data.Get("enabled")),
		Endpoint:      toString(data.Get("endpoint")),
		IncludeRuns:   toBool(data.Get("include_runs")),
		CustomHeaders: toOptionalStringMap(data.Get("custom_headers")),
	}

	if retryOnFailure, ok := data.GetOk("retry_on_failure"); ok {
		input.RetryOnFailure = toOptionalBool(retryOnFailure)
	}

	variables := map[string]interface{}{
		"input": input,
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	return resourceAuditTrailWebhookRead(ctx, data, i)
}

func resourceAuditTrailWebhookDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		AuditTrailWebhook *structs.AuditTrailWebhookRead `graphql:"auditTrailDeleteWebhook"`
	}
	if err := i.(*internal.Client).Mutate(ctx, "AuditTrailWebhookDelete", &mutation, nil); err != nil {
		return diag.Errorf("could not delete audit trail webhook: %v", internal.FromSpaceliftError(err))
	}

	data.SetId("")

	return nil
}
