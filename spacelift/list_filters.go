package spacelift

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func listFilterAttributes(plural string) map[string]listschema.Attribute {
	return map[string]listschema.Attribute{
		"name": listschema.StringAttribute{
			Description: fmt.Sprintf("Only list %s whose name matches this pattern. `*` matches any characters, and case is ignored.", plural),
			Optional:    true,
		},
		"space_id": listschema.StringAttribute{
			Description: fmt.Sprintf("Only list %s in this space.", plural),
			Optional:    true,
		},
		"include_child_spaces": listschema.BoolAttribute{
			Description: fmt.Sprintf("With `space_id`, also list %s in its child spaces, at any depth.", plural),
			Optional:    true,
		},
		"labels": listschema.ListAttribute{
			Description: fmt.Sprintf("Only list %s that have all of these labels.", plural),
			ElementType: types.StringType,
			Optional:    true,
		},
	}
}

// listFilter is a list resource's filter config, ready to use either as search
// API predicates or to match objects in the provider.
type listFilter struct {
	name string
	// name as a regexp, for matching in the provider.
	namePattern        *regexp.Regexp
	spaceID            string
	includeChildSpaces bool
	labels             []string
	// Loaded only when filtering by space, which needs the space hierarchy.
	spaces spaceTree
}

func newListFilter(ctx context.Context, client *internal.Client, config tfsdk.Config, spaceAttribute string) (*listFilter, fwdiag.Diagnostics) {
	var name, spaceID types.String
	var includeChildSpaces types.Bool
	var labels types.List

	var diags fwdiag.Diagnostics
	diags.Append(config.GetAttribute(ctx, fwpath.Root("name"), &name)...)
	diags.Append(config.GetAttribute(ctx, fwpath.Root(spaceAttribute), &spaceID)...)
	diags.Append(config.GetAttribute(ctx, fwpath.Root("include_child_spaces"), &includeChildSpaces)...)
	diags.Append(config.GetAttribute(ctx, fwpath.Root("labels"), &labels)...)
	if diags.HasError() {
		return nil, diags
	}

	filter := &listFilter{
		name:               name.ValueString(),
		namePattern:        namePattern(name.ValueString()),
		spaceID:            spaceID.ValueString(),
		includeChildSpaces: includeChildSpaces.ValueBool(),
	}

	if filter.includeChildSpaces && filter.spaceID == "" {
		diags.AddAttributeError(fwpath.Root("include_child_spaces"), "missing "+spaceAttribute, fmt.Sprintf("include_child_spaces only works with %s", spaceAttribute))
		return nil, diags
	}

	if diags.Append(labels.ElementsAs(ctx, &filter.labels, false)...); diags.HasError() {
		return nil, diags
	}

	if filter.spaceID != "" {
		spaces, err := loadSpaceTree(ctx, client)
		if err != nil {
			diags.AddError("could not load spaces", err.Error())
			return nil, diags
		}
		if _, ok := spaces[filter.spaceID]; !ok {
			diags.AddAttributeError(fwpath.Root(spaceAttribute), "space not found", fmt.Sprintf("space %q does not exist or you can't read it", filter.spaceID))
			return nil, diags
		}
		filter.spaces = spaces
	}

	return filter, diags
}

// predicates returns the search API predicates for the filter. Objects must
// match all of them.
func (f *listFilter) predicates() []search.SearchQueryPredicate {
	var predicates []search.SearchQueryPredicate

	if f.name != "" {
		predicates = append(predicates, namePredicate(f.name))
	}

	if f.spaceID != "" {
		predicates = append(predicates, spacePredicate(f.spaces.path(f.spaceID), f.includeChildSpaces))
	}

	// One predicate per label, since a single one matches any of its values.
	for _, label := range f.labels {
		predicates = append(predicates, search.SearchQueryPredicate{
			Field:      "label",
			Constraint: search.SearchQueryFieldConstraint{StringEquals: &[]graphql.String{graphql.String(label)}},
		})
	}

	return predicates
}

// matches reports whether an object passes the filter, for objects that can't
// be searched.
func (f *listFilter) matches(name, spaceID string, labels []string) bool {
	if !f.namePattern.MatchString(name) {
		return false
	}

	if f.spaceID != "" && spaceID != f.spaceID && (!f.includeChildSpaces || !f.spaces.isDescendant(spaceID, f.spaceID)) {
		return false
	}

	for _, label := range f.labels {
		if !slices.Contains(labels, label) {
			return false
		}
	}

	return true
}

// namePattern matches names the way the search API does: the whole name, case
// insensitive, with * standing for any characters.
func namePattern(pattern string) *regexp.Regexp {
	if pattern == "" {
		return regexp.MustCompile("")
	}

	parts := strings.Split(pattern, "*")
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}

	return regexp.MustCompile("(?is)^" + strings.Join(parts, ".*") + "$")
}

func namePredicate(pattern string) search.SearchQueryPredicate {
	return search.SearchQueryPredicate{
		Field:      "name",
		Constraint: search.SearchQueryFieldConstraint{StringMatches: &[]graphql.String{graphql.String(pattern)}},
	}
}

func spacePredicate(path string, includeDescendants bool) search.SearchQueryPredicate {
	return search.SearchQueryPredicate{
		Field: "space",
		Constraint: search.SearchQueryFieldConstraint{HierarchyNode: &search.HierarchyNodeList{
			Nodes: []search.HierarchyNode{{Equals: graphql.String(path), IncludeDescendants: graphql.Boolean(includeDescendants)}},
		}},
	}
}

// spaceTree maps each space ID to its parent's, with "" for root.
type spaceTree map[string]string

func loadSpaceTree(ctx context.Context, client *internal.Client) (spaceTree, error) {
	var query struct {
		Spaces []struct {
			ID          string  `graphql:"id"`
			ParentSpace *string `graphql:"parentSpace"`
		} `graphql:"spaces"`
	}

	if err := client.Query(ctx, "ListSpaceTree", &query, nil); err != nil {
		return nil, errors.Wrap(err, "could not query spaces")
	}

	tree := make(spaceTree, len(query.Spaces))
	for _, space := range query.Spaces {
		tree[space.ID] = ""
		if space.ParentSpace != nil {
			tree[space.ID] = *space.ParentSpace
		}
	}

	return tree, nil
}

// path returns the space's path as the search API matches it: the IDs from
// root down to the space, e.g. root/team-01H.../app-01H....
func (t spaceTree) path(spaceID string) string {
	var ids []string
	// Bounded so a malformed tree can't loop forever.
	for id := spaceID; id != "" && len(ids) <= len(t); id = t[id] {
		ids = append(ids, id)
	}
	slices.Reverse(ids)

	return strings.Join(ids, "/")
}

func (t spaceTree) isDescendant(spaceID, ancestorID string) bool {
	for id, depth := t[spaceID], 0; id != "" && depth <= len(t); id, depth = t[id], depth+1 {
		if id == ancestorID {
			return true
		}
	}

	return false
}
