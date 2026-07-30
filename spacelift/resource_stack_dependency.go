package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oklog/ulid/v2"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// The Framework picks up Configure and ImportState by asserting on the resource at
// runtime, so a signature that drifts out of line disables the capability instead of
// breaking the build: Configure never runs and r.client stays nil, or the resource
// quietly stops supporting import. These assertions turn that into a compile error
// naming the signature it expected.
var (
	_ resource.Resource                = (*stackDependencyResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackDependencyResource)(nil)
	_ resource.ResourceWithImportState = (*stackDependencyResource)(nil)
)

// NewStackDependencyResource returns the Plugin Framework implementation of
// spacelift_stack_dependency.
func NewStackDependencyResource() resource.Resource { return &stackDependencyResource{} }

type stackDependencyResource struct {
	client *internal.Client
}

type stackDependencyModel struct {
	ID               types.String `tfsdk:"id"`
	StackID          types.String `tfsdk:"stack_id"`
	DependsOnStackID types.String `tfsdk:"depends_on_stack_id"`
}

func (r *stackDependencyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "spacelift_stack_dependency"
}

func (r *stackDependencyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// ProviderData is nil during schema-validation walks.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*internal.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"unexpected provider data",
			fmt.Sprintf("expected *internal.Client, got %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *stackDependencyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "" +
			"`spacelift_stack_dependency` represents a Spacelift **stack dependency** - " +
			"a dependency between two stacks. When one stack depends on another, the tracked runs " +
			"of the stack will not start until the dependent stack is successfully finished. Additionally, " +
			"changes to the dependency will trigger the dependent.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_id": schema.StringAttribute{
				Description: "immutable ID (slug) of stack which has a dependency.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"depends_on_stack_id": schema.StringAttribute{
				Description: "immutable ID (slug) of stack to depend on.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *stackDependencyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stackDependencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		StackDependency structs.StackDependency `graphql:"stackDependencyCreate(input: $input)"`
	}

	variables := map[string]any{
		"input": structs.StackDependencyInput{
			StackID:          toID(plan.StackID.ValueString()),
			DependsOnStackID: toID(plan.DependsOnStackID.ValueString()),
		},
	}

	if err := r.client.Mutate(ctx, "StackDependencyCreate", &mutation, variables); err != nil {
		resp.Diagnostics.AddError("could not create stack dependency", err.Error())
		return
	}

	dep := mutation.StackDependency
	plan.ID = types.StringValue(path.Join(dep.Stack.ID, dep.ID))
	plan.StackID = types.StringValue(dep.Stack.ID)
	plan.DependsOnStackID = types.StringValue(dep.DependsOnStack.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackDependencyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stackDependencyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dependency, err := r.fetchByStackID(
		ctx,
		toID(state.StackID.ValueString()),
		toID(state.DependsOnStackID.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError("could not query for stack dependency", err.Error())
		return
	}

	if dependency == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(path.Join(dependency.Stack.ID, dependency.ID))
	state.StackID = types.StringValue(dependency.Stack.ID)
	state.DependsOnStackID = types.StringValue(dependency.DependsOnStack.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update exists only to satisfy resource.Resource. Both attributes force replacement,
// so there is no in-place update path and the SDKv2 implementation had no UpdateContext.
func (r *stackDependencyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stackDependencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackDependencyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stackDependencyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, depID, err := parseStackDependencyID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("could not parse stack dependency ID", err.Error())
		return
	}

	var mutation struct {
		StackDependency *structs.StackDependency `graphql:"stackDependencyDelete(id: $id)"`
	}

	variables := map[string]any{"id": graphql.ID(depID)}

	if err := r.client.Mutate(ctx, "StackDependencyDelete", &mutation, variables); err != nil {
		resp.Diagnostics.AddError("could not delete stack dependency", err.Error())
	}
}

// ImportState accepts both supported ID formats: the legacy "stackID/dependencyULID"
// and the human-readable "stackID/dependsOnStackID". A successful ULID parse of the
// second part selects the legacy lookup.
func (r *stackDependencyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	stackID, secondPart, err := parseStackDependencyID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("invalid import ID format", err.Error())
		return
	}

	var dependency *structs.StackDependency

	if _, ulidErr := ulid.Parse(secondPart); ulidErr == nil {
		dependency, err = r.fetchByID(ctx, toID(stackID), toID(secondPart))
	} else {
		dependency, err = r.fetchByStackID(ctx, toID(stackID), toID(secondPart))
	}

	if err != nil {
		resp.Diagnostics.AddError("could not query for stack dependency", err.Error())
		return
	}

	if dependency == nil {
		resp.Diagnostics.AddError(
			"stack dependency not found",
			fmt.Sprintf("no stack dependency found for import ID: %s", path.Join(stackID, secondPart)),
		)
		return
	}

	state := stackDependencyModel{
		ID:               types.StringValue(path.Join(dependency.Stack.ID, dependency.ID)),
		StackID:          types.StringValue(dependency.Stack.ID),
		DependsOnStackID: types.StringValue(dependency.DependsOnStack.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// fetchByID looks a dependency up by its ULID (legacy import format).
func (r *stackDependencyResource) fetchByID(ctx context.Context, stackID, dependencyID graphql.ID) (*structs.StackDependency, error) {
	return r.fetch(ctx, map[string]any{
		"id":               dependencyID,
		"stackId":          stackID,
		"dependsOnStackId": toID(""),
	})
}

// fetchByStackID looks a dependency up by the stack it depends on.
func (r *stackDependencyResource) fetchByStackID(ctx context.Context, stackID, dependsOnStackID graphql.ID) (*structs.StackDependency, error) {
	return r.fetch(ctx, map[string]any{
		"id":               toID(""),
		"stackId":          stackID,
		"dependsOnStackId": dependsOnStackID,
	})
}

func (r *stackDependencyResource) fetch(ctx context.Context, variables map[string]any) (*structs.StackDependency, error) {
	var query struct {
		Stack *struct {
			Dependency *structs.StackDependency `graphql:"dependency(id: $id, dependsOnStackId: $dependsOnStackId)"`
		} `graphql:"stack(id: $stackId)"`
	}

	if err := r.client.Query(ctx, "StackDependencyRead", &query, variables); err != nil {
		return nil, err
	}

	if query.Stack == nil {
		return nil, nil
	}

	return query.Stack.Dependency, nil
}

func parseStackDependencyID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unexpected resource ID format: %s", id)
	}

	return idParts[0], idParts[1], nil
}
