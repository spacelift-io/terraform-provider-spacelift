package spacelift

import (
	"context"
	"fmt"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

const listPageSize = 50

// sdkListResource lists the objects of an SDKv2 resource type. T is the API
// object the resource's populate function takes.
type sdkListResource[T any] struct {
	typeName string
	singular string
	plural   string
	resource func() *schema.Resource
	// fetchPage returns one page of objects. Types without a search query return
	// everything and an empty PageInfo.
	fetchPage func(ctx context.Context, client *internal.Client, input search.SearchInput) ([]T, search.PageInfo, error)
	populate  func(d *schema.ResourceData, object *T) diag.Diagnostics
	name      func(object *T) string

	// Optional. keep drops objects the search could not filter out.
	keep func(filter *listFilter, object *T) bool
	// Optional. Defaults to filter.predicates().
	predicates func(ctx context.Context, config tfsdk.Config, filter *listFilter) ([]search.SearchQueryPredicate, fwdiag.Diagnostics)
	// Optional. Defaults to listFilterAttributes(plural).
	attributes func() map[string]listschema.Attribute
	// Optional. The config attribute holding the space filter, "space_id" by default.
	spaceAttribute string

	client *internal.Client
}

var (
	_ list.ListResourceWithRawV6Schemas = &sdkListResource[struct{}]{}
	_ list.ListResourceWithConfigure    = &sdkListResource[struct{}]{}
)

func (l *sdkListResource[T]) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = l.typeName
}

func (l *sdkListResource[T]) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	attributes := listFilterAttributes(l.plural)
	if l.attributes != nil {
		attributes = l.attributes()
	}

	resp.Schema = listschema.Schema{
		Description: fmt.Sprintf("Lists every %s in the account.", l.singular),
		Attributes:  attributes,
	}
}

// The framework server speaks protocol 6 and ignores RawV5Schemas, so the SDKv2
// schemas are translated.
func (l *sdkListResource[T]) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	r := l.resource()
	resp.ProtoV6Schema = translate.Schema(r.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(r.ProtoIdentitySchema(ctx)())
}

func (l *sdkListResource[T]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// ProviderData is nil during schema-validation walks.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*internal.Client)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data", fmt.Sprintf("expected *internal.Client, got %T", req.ProviderData))
		return
	}
	l.client = client
}

func (l *sdkListResource[T]) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	errorSummary := "could not list " + l.plural

	if l.client == nil {
		var diags fwdiag.Diagnostics
		diags.AddError(errorSummary, "provider is not configured")
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	predicates, filter, diags := l.filter(ctx, req.Config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	r := l.resource()

	stream.Results = func(push func(list.ListResult) bool) {
		input := search.SearchInput{First: graphql.NewInt(listPageSize), Predicates: &predicates}
		var count int64

		for {
			objects, pageInfo, err := l.fetchPage(ctx, l.client, input)
			if err != nil {
				result := req.NewListResult(ctx)
				result.Diagnostics.AddError(errorSummary, err.Error())
				push(result)
				return
			}

			for i := range objects {
				object := &objects[i]
				if l.keep != nil && !l.keep(filter, object) {
					continue
				}

				result := sdkListResult(ctx, req, r, l.name(object), func(d *schema.ResourceData) diag.Diagnostics {
					return l.populate(d, object)
				})
				if !push(result) || result.Diagnostics.HasError() {
					return
				}

				if count++; req.Limit > 0 && count >= req.Limit {
					return
				}
			}

			if !pageInfo.HasNextPage {
				return
			}

			after := graphql.String(pageInfo.EndCursor)
			input.After = &after
		}
	}
}

func (l *sdkListResource[T]) filter(ctx context.Context, config tfsdk.Config) ([]search.SearchQueryPredicate, *listFilter, fwdiag.Diagnostics) {
	spaceAttribute := l.spaceAttribute
	if spaceAttribute == "" {
		spaceAttribute = "space_id"
	}

	filter, diags := newListFilter(ctx, l.client, config, spaceAttribute)
	if diags.HasError() {
		return nil, nil, diags
	}

	if l.predicates == nil {
		return filter.predicates(), filter, diags
	}

	predicates, diags := l.predicates(ctx, config, filter)
	return predicates, filter, diags
}

func sdkListResult(ctx context.Context, req list.ListRequest, r *schema.Resource, displayName string, populate func(d *schema.ResourceData) diag.Diagnostics) list.ListResult {
	result := req.NewListResult(ctx)

	d := r.Data(&terraform.InstanceState{})
	// Populate warnings are about the resource's config, which a list does not have.
	for _, sdkDiag := range populate(d) {
		if sdkDiag.Severity == diag.Error {
			result.Diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
	}
	if result.Diagnostics.HasError() {
		return result
	}

	result.DisplayName = displayName
	if result.DisplayName == "" {
		result.DisplayName = d.Id()
	}

	identity, err := d.TfTypeIdentityState()
	if err != nil {
		result.Diagnostics.AddError("could not convert resource identity", err.Error())
		return result
	}
	result.Identity.Raw = *identity

	if req.IncludeResource {
		state, err := d.TfTypeResourceState()
		if err != nil {
			result.Diagnostics.AddError("could not convert resource state", err.Error())
			return result
		}
		result.Resource.Raw = *state
	}

	return result
}

// searchPage is the output of a search query, e.g. searchStacks.
type searchPage[T any] struct {
	Edges []struct {
		Node T `graphql:"node"`
	} `graphql:"edges"`
	PageInfo search.PageInfo `graphql:"pageInfo"`
}

func (p searchPage[T]) nodes() []T {
	nodes := make([]T, len(p.Edges))
	for i, edge := range p.Edges {
		nodes[i] = edge.Node
	}
	return nodes
}

// An empty identity would match every other empty one.
func hasULID(ulid string) bool { return ulid != "" }
