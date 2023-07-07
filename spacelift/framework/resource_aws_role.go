package framework

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	slGQL "github.com/spacelift-io/terraform-provider-spacelift/spacelift/graphql"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func NewAWSRoleResource() resource.Resource {
	return &AWSRoleResource{}
}

type AWSRoleResource struct {
	Client *internal.Client
}

// AWSRoleResourceModel describes the Terraform resource data model to match the resource schema.
type AWSRoleResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	ModuleID                    types.String `tfsdk:"module_id"`
	StackID                     types.String `tfsdk:"stack_id"`
	RoleARN                     types.String `tfsdk:"role_arn"`
	GenerateCredentialsInWorker types.Bool   `tfsdk:"generate_credentials_in_worker"`
	ExternalID                  types.String `tfsdk:"external_id"`
	DurationSeconds             types.Int64  `tfsdk:"duration_seconds"`
}

func (r *AWSRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*internal.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *internal.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.Client = client
}

func (r *AWSRoleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "spacelift_aws_role"
}

func (r *AWSRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := AWSRoleResourceModel{}

	if err := importIntegration(ctx, req.ID, &state); err != nil {
		resp.Diagnostics.AddError("could not import AWS role", err.Error())

		return
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AWSRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "" +
			"**NOTE:** while this resource continues to work, we have replaced it with the `spacelift_aws_integration` " +
			"resource. The new resource allows integrations to be shared by multiple stacks/modules " +
			"and also supports separate read vs write roles. Please use the `spacelift_aws_integration` resource instead.\n\n" +
			"`spacelift_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) " +
			"between the Spacelift worker and an individual stack or module. " +
			"If this is set, Spacelift will use AWS STS to assume the supplied IAM role and " +
			"put its temporary credentials in the runtime environment." +
			"\n\n" +
			"If you use private workers, you can also assume IAM role on the worker side using " +
			"your own AWS credentials (e.g. from EC2 instance profile)." +
			"\n\n" +
			"Note: when assuming credentials for **shared worker**, Spacelift will use `$accountName@$stackID` " +
			"or `$accountName@$moduleID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) " +
			"and `$runID@$stackID@$accountName` truncated to 64 characters as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the module or stack which assumes the AWS IAM role",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"module_id": schema.StringAttribute{
				Description: "ID of the module which assumes the AWS IAM role",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				Description: "ARN of the AWS IAM role to attach",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"stack_id": schema.StringAttribute{
				Description: "ID of the stack which assumes the AWS IAM role",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"generate_credentials_in_worker": schema.BoolAttribute{
				Description: "Generate AWS credentials in the private worker. Defaults to `false`.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"external_id": schema.StringAttribute{
				Description: "Custom external ID (works only for private workers).",
				Optional:    true,
			},
			"duration_seconds": schema.Int64Attribute{
				Description: "AWS IAM role session duration in seconds",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

func (r *AWSRoleResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("module_id"),
			path.MatchRoot("stack_id"),
		),
	}
}

func (r *AWSRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AWSRoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ID string
	if data.StackID.IsNull() {
		if err := verifyStack(ctx, data.StackID.ValueString(), r.Client); err != nil {
			resp.Diagnostics.AddError("Could not verify stack ID", err.Error())

			return
		}

		ID = data.StackID.ValueString()
	} else {
		if err := verifyModule(ctx, data.ModuleID.ValueString(), r.Client); err != nil {
			resp.Diagnostics.AddError("Could not verify module ID", err.Error())

			return
		}

		ID = data.ModuleID.ValueString()
	}

	var err error

	for i := 0; i < 5; i++ {
		err = r.Set(ctx, ID, data)
		if err == nil || !strings.Contains(err.Error(), "could not assume") || i == 4 {
			break
		}

		// Yay for eventual consistency.
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		resp.Diagnostics.AddError("could not create AWS role delegation", err.Error())

		return
	}

	data.ID = types.StringValue(ID)

	if err := resourceAWSRoleRead(ctx, data, r.Client); err != nil {
		resp.Diagnostics.AddError("could not read AWS role", err.Error())
	}

	return
}

func resourceAWSRoleRead(ctx context.Context, d *AWSRoleResourceModel, client *internal.Client) error {
	if !d.ModuleID.IsNull() {
		return resourceModuleAWSRoleRead(ctx, d, client)
	}

	return resourceStackAWSRoleRead(ctx, d, client)
}

func resourceModuleAWSRoleRead(ctx context.Context, d *AWSRoleResourceModel, client *internal.Client) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.ID.ValueString())}

	if err := client.Query(ctx, "ModuleAWSRoleRead", &query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	if query.Module == nil {
		d.ID = types.StringNull()

		return nil
	}

	d.SetIntegration(&query.Module.Integrations)

	return nil
}

func resourceStackAWSRoleRead(ctx context.Context, d *AWSRoleResourceModel, client *internal.Client) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.ID.ValueString())}

	if err := client.Query(ctx, "StackAWSRoleRead", &query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	if query.Stack == nil {
		d.ID = types.StringNull()

		return nil
	}

	d.SetIntegration(query.Stack.Integrations)

	return nil
}

func (r *AWSRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AWSRoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if !data.ModuleID.IsNull() {
		if err := resourceModuleAWSRoleRead(ctx, &data, r.Client); err != nil {
			resp.Diagnostics.AddError("could not read AWS role", err.Error())

			return
		}
	}

	if err := resourceStackAWSRoleRead(ctx, &data, r.Client); err != nil {
		resp.Diagnostics.AddError("could not read AWS role", err.Error())

		return
	}

	return
}

func (r *AWSRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AWSRoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var ID string
	if !data.ModuleID.IsNull() {
		ID = data.ModuleID.ValueString()
	} else {
		ID = data.StackID.ValueString()
	}

	if err := r.Set(ctx, ID, &data); err != nil {
		resp.Diagnostics.AddError("could not update AWS role", err.Error())

		return
	}

	if err := resourceAWSRoleRead(ctx, &data, r.Client); err != nil {
		resp.Diagnostics.AddError("could not read AWS role", err.Error())

		return
	}

	return
}

func (r *AWSRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AWSRoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsDelete(id: $id)"`
	}

	variables := map[string]interface{}{}

	if !data.ModuleID.IsNull() {
		variables["id"] = data.ModuleID.ValueString()
	} else {
		variables["id"] = data.StackID.ValueString()
	}

	if err := r.Client.Mutate(ctx, "AWSRoleDelete", &mutation, variables); err != nil {
		resp.Diagnostics.AddError("could not delete AWS role delegation", err.Error())

		return
	}

	if mutation.AttachAWSRole.Activated {
		resp.Diagnostics.AddError("did not disable AWS integration, still reporting as activated", "")

		return
	}
}

func (r *AWSRoleResource) Set(ctx context.Context, id string, d *AWSRoleResourceModel) error {
	var mutation struct {
		AttachAWSRole struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationAwsCreate(id: $id, roleArn: $roleArn, generateCredentialsInWorker: $generateCredentialsInWorker, externalID: $externalID, durationSeconds: $durationSeconds)"`
	}

	variables := map[string]interface{}{
		"id":                          graphql.ID(id),
		"roleArn":                     graphql.String(d.RoleARN.ValueString()),
		"generateCredentialsInWorker": graphql.Boolean(d.GenerateCredentialsInWorker.ValueBool()),
	}

	if !d.ExternalID.IsNull() {
		variables["externalID"] = slGQL.ToOptionalString(d.ExternalID.ValueString())
	} else {
		variables["externalID"] = (*graphql.String)(nil)
	}

	if !d.DurationSeconds.IsNull() {
		variables["durationSeconds"] = slGQL.ToOptionalString(d.DurationSeconds.ValueInt64())
	} else {
		variables["durationSeconds"] = (*graphql.Int)(nil)
	}

	if err := r.Client.Mutate(ctx, "AWSRoleSet", &mutation, variables); err != nil {
		return errors.Wrap(err, "could not set AWS role delegation")
	}

	if !mutation.AttachAWSRole.Activated {
		return errors.New("AWS integration not activated")
	}

	return nil
}

func (d *AWSRoleResourceModel) SetIntegration(integrations *structs.Integrations) {
	if roleARN := integrations.AWS.AssumedRoleARN; roleARN != nil {
		d.RoleARN = types.StringPointerValue(roleARN)
	} else {
		d.RoleARN = types.StringNull()
	}

	d.GenerateCredentialsInWorker = types.BoolValue(integrations.AWS.GenerateCredentialsInWorker)
	d.ExternalID = types.StringPointerValue(integrations.AWS.ExternalID)
	d.DurationSeconds = types.Int64Value(int64(*integrations.AWS.DurationSeconds))
}
