package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ignoreOnceCreatedModifier suppresses diffs for a string attribute once the
// resource has been created. Mirrors SDKv2's ignoreOnceCreated DiffSuppressFunc.
//
// Behaviour:
//   - On create (state is null): allow the planned value through unchanged.
//   - On update (state exists): replace planned value with state value, preventing
//     Terraform from showing a diff even if the config value differs from state.
//
// Used for the deprecated `secret` field on webhook resources, where Read always
// writes "" to state and the API never returns the real value.
type ignoreOnceCreatedModifier struct{}

func (m ignoreOnceCreatedModifier) Description(_ context.Context) string {
	return "Once the resource is created, changes to this attribute are ignored."
}

func (m ignoreOnceCreatedModifier) MarkdownDescription(_ context.Context) string {
	return "Once the resource is created, changes to this attribute are ignored."
}

func (m ignoreOnceCreatedModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// req.State.Raw.IsNull() is true when the resource does not yet exist.
	if req.State.Raw.IsNull() {
		return
	}
	// Resource exists: always use state value to suppress diffs.
	resp.PlanValue = req.StateValue
}
