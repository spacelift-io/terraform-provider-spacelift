package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

type namedWebhookResource struct{ client *internal.Client }

type namedWebhookModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Endpoint        types.String `tfsdk:"endpoint"`
	SpaceID         types.String `tfsdk:"space_id"`
	Name            types.String `tfsdk:"name"`
	Labels          types.Set    `tfsdk:"labels"`
	Secret          types.String `tfsdk:"secret"`
	SecretWo        types.String `tfsdk:"secret_wo"`
	SecretWoVersion types.String `tfsdk:"secret_wo_version"`
	RetryOnFailure  types.Bool   `tfsdk:"retry_on_failure"`
}

func NewNamedWebhookResource() resource.Resource { return &namedWebhookResource{} }

func (r *namedWebhookResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "spacelift_named_webhook"
}

func (r *namedWebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`spacelift_named_webhook` represents a named webhook endpoint used for creating webhooks" +
			"which are referred to in Notification policies to route messages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "enables or disables sending webhooks.",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "endpoint to send the requests to",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"space_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the space the webhook is in",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "the name for the webhook which will also be used to generate the id",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"labels": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "labels for the webhook to use when referring in policies or filtering them",
			},
			"secret": schema.StringAttribute{
				Optional:           true,
				Sensitive:          true,
				DeprecationMessage: "`secret` is deprecated. Please use secret_wo in combination with secret_wo_version",
				Description:        "secret used to sign each request so you're able to verify that the request comes from us. Defaults to an empty value. Note that once it's created, it will be just an empty string in the state due to security reasons.",
				PlanModifiers: []planmodifier.String{
					ignoreOnceCreatedModifier{},
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("secret_wo"),
						path.MatchRoot("secret_wo_version"),
					),
				},
			},
			"secret_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "secret used to sign each request so you're able to verify that the request comes from us. Defaults to an empty value. The secret is not stored in the state. Modify secret_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("secret")),
					stringvalidator.AlsoRequires(path.MatchRoot("secret_wo_version")),
				},
			},
			"secret_wo_version": schema.StringAttribute{
				Optional:    true,
				Description: "Used together with secret_wo to trigger an update to the secret. Increment this value when an update to secret_wo is required. This field requires Terraform/OpenTofu 1.11+.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("secret")),
					stringvalidator.AlsoRequires(path.MatchRoot("secret_wo")),
				},
			},
			"retry_on_failure": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "whether to retry the webhook in case of failure. Defaults to `false`.",
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *namedWebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*internal.Client)
}

func (r *namedWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namedWebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := namedWebhookExtractSecret(plan)

	labels := make([]graphql.String, 0)
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		for _, label := range plan.Labels.Elements() {
			labels = append(labels, graphql.String(label.(types.String).ValueString()))
		}
	}

	input := structs.NamedWebhooksIntegrationInput{
		Enabled:  graphql.Boolean(plan.Enabled.ValueBool()),
		Endpoint: graphql.String(plan.Endpoint.ValueString()),
		Space:    graphql.ID(plan.SpaceID.ValueString()),
		Name:     graphql.String(plan.Name.ValueString()),
		Secret:   toOptionalString(secret),
		Labels:   labels,
	}

	if !plan.RetryOnFailure.IsNull() && !plan.RetryOnFailure.IsUnknown() {
		input.RetryOnFailure = toOptionalBool(plan.RetryOnFailure.ValueBool())
	}

	var mutation struct {
		WebhooksIntegration struct {
			ID      string `graphql:"id"`
			Enabled bool   `graphql:"enabled"`
		} `graphql:"namedWebhooksIntegrationCreate(input: $input)"`
	}

	if err := r.client.Mutate(ctx, "NamedWebhookCreate", &mutation, map[string]any{"input": input}); err != nil {
		resp.Diagnostics.AddError("could not create named webhook", internal.FromSpaceliftError(err).Error())
		return
	}

	plan.ID = types.StringValue(mutation.WebhooksIntegration.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readInto(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namedWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namedWebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readInto(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		// Resource was deleted outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *namedWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan namedWebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := make([]graphql.String, 0)
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		for _, label := range plan.Labels.Elements() {
			labels = append(labels, graphql.String(label.(types.String).ValueString()))
		}
	}

	input := structs.NamedWebhooksIntegrationInput{
		Enabled:  graphql.Boolean(plan.Enabled.ValueBool()),
		Endpoint: graphql.String(plan.Endpoint.ValueString()),
		Space:    graphql.ID(plan.SpaceID.ValueString()),
		Name:     graphql.String(plan.Name.ValueString()),
		Labels:   labels,
	}

	// Only send secret if explicitly provided — prevents accidental override on unrelated updates.
	secret := namedWebhookExtractSecret(plan)
	if secret != "" {
		input.Secret = toOptionalString(secret)
	}

	if !plan.RetryOnFailure.IsNull() && !plan.RetryOnFailure.IsUnknown() {
		input.RetryOnFailure = toOptionalBool(plan.RetryOnFailure.ValueBool())
	}

	var mutation struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegrationUpdate(id: $id, input: $input)"`
	}

	if err := r.client.Mutate(ctx, "NamedWebhookUpdate", &mutation, map[string]any{
		"id":    toID(plan.ID.ValueString()),
		"input": input,
	}); err != nil {
		resp.Diagnostics.AddError("could not update named webhook", internal.FromSpaceliftError(err).Error())
		return
	}

	r.readInto(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namedWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namedWebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		WebhooksIntegration struct {
			ID string `graphql:"id"`
		} `graphql:"namedWebhooksIntegrationDelete(id: $webhook)"`
	}

	if err := r.client.Mutate(ctx, "NamedWebhookDelete", &mutation, map[string]any{
		"webhook": toID(state.ID.ValueString()),
	}); err != nil {
		resp.Diagnostics.AddError("could not delete named webhook", internal.FromSpaceliftError(err).Error())
	}
}

func (r *namedWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readInto queries the API and populates model in place.
// Sets model.ID to null when the webhook no longer exists (caller should call
// resp.State.RemoveResource if ID is null after this returns).
func (r *namedWebhookResource) readInto(ctx context.Context, id string, model *namedWebhookModel, diags *diag.Diagnostics) {
	var query struct {
		Webhook *structs.NamedWebhooksIntegration `graphql:"namedWebhooksIntegration(id: $id)"`
	}

	if err := r.client.Query(ctx, "GetNamedWebhook", &query, map[string]any{"id": graphql.ID(id)}); err != nil {
		diags.AddError("could not query named webhook", err.Error())
		return
	}

	if query.Webhook == nil {
		model.ID = types.StringNull()
		return
	}

	wh := query.Webhook
	model.ID = types.StringValue(wh.ID)
	model.Name = types.StringValue(wh.Name)
	model.Endpoint = types.StringValue(wh.Endpoint)
	model.Enabled = types.BoolValue(wh.Enabled)
	model.SpaceID = types.StringValue(wh.Space.ID)
	model.Secret = types.StringValue("") // API never returns the real secret value.
	if wh.RetryOnFailure != nil {
		model.RetryOnFailure = types.BoolValue(*wh.RetryOnFailure)
	} else {
		model.RetryOnFailure = types.BoolValue(false)
	}

	labelValues := make([]attr.Value, len(wh.Labels))
	for i, label := range wh.Labels {
		labelValues[i] = types.StringValue(label)
	}
	labelsSet, d := types.SetValue(types.StringType, labelValues)
	diags.Append(d...)
	model.Labels = labelsSet
}

// namedWebhookExtractSecret returns the secret value from either the deprecated
// `secret` field or the write-only `secret_wo` field, mirroring SDKv2's
// internal.ExtractWriteOnlyField("secret", "secret_wo", "secret_wo_version", d).
func namedWebhookExtractSecret(plan namedWebhookModel) string {
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() && plan.Secret.ValueString() != "" {
		return plan.Secret.ValueString()
	}
	if !plan.SecretWo.IsNull() && !plan.SecretWo.IsUnknown() {
		return plan.SecretWo.ValueString()
	}
	return ""
}
