package spacelift

import (
	"context"
	"fmt"
	"slices"
	"strings"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func NewPolicyListResource() list.ListResource {
	return &sdkListResource[structs.Policy]{
		typeName: "spacelift_policy", singular: "policy", plural: "policies",
		resource: resourcePolicy,
		attributes: func() map[string]listschema.Attribute {
			attributes := listFilterAttributes("policies")
			attributes["type"] = listschema.StringAttribute{
				Description: "Only list policies of this type, e.g. `PLAN` or `GIT_PUSH`.",
				Optional:    true,
			}
			return attributes
		},
		predicates: func(ctx context.Context, config tfsdk.Config, filter *listFilter) ([]search.SearchQueryPredicate, fwdiag.Diagnostics) {
			var policyType types.String
			diags := config.GetAttribute(ctx, fwpath.Root("type"), &policyType)
			// The API ignores unknown types and returns nothing.
			if t := policyType.ValueString(); t != "" && !slices.Contains(policyTypes, t) {
				diags.AddAttributeError(fwpath.Root("type"), "invalid policy type", fmt.Sprintf("%q is not one of %s", t, strings.Join(policyTypes, ", ")))
			}
			if diags.HasError() {
				return nil, diags
			}

			predicates := filter.predicates()
			if policyType.ValueString() != "" {
				predicates = append(predicates, search.SearchQueryPredicate{
					Field:      "type",
					Constraint: search.SearchQueryFieldConstraint{EnumEquals: &[]graphql.String{graphql.String(policyType.ValueString())}},
				})
			}
			return predicates, diags
		},
		fetchPage: func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]structs.Policy, search.PageInfo, error) {
			var query struct {
				Page searchPage[structs.Policy] `graphql:"searchPolicies(input: $input)"`
			}
			err := client.Query(ctx, "PolicyList", &query, map[string]any{"input": input})
			return query.Page.nodes(), query.Page.PageInfo, err
		},
		keep:     func(_ *listFilter, policy *structs.Policy) bool { return hasULID(policy.ULID) },
		name:     func(policy *structs.Policy) string { return policy.Name },
		populate: populatePolicy,
	}
}
