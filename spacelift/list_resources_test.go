package spacelift

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type node = map[string]any

// TestListResources lists every type through the muxed provider against a fake
// API. Search queries are served in two pages.
func TestListResources(t *testing.T) {
	with := func(extra node, nodes ...node) []node {
		for _, n := range nodes {
			n["name"] = "Name " + n["id"].(string)
			n["space"] = "root"
			for k, v := range extra {
				n[k] = v
			}
		}
		return nodes
	}
	// The empty ULID must be skipped.
	ulids := func(extra node) []node {
		return with(extra, node{"id": "old", "ulid": ""}, node{"id": "one", "ulid": "U1"}, node{"id": "two", "ulid": "U2"})
	}
	ids := func() []node { return with(nil, node{"id": "A"}, node{"id": "B"}) }
	// Populating a stack or module needs the objects the API always returns.
	stack := node{"integrations": node{}, "vendorConfig": node{"__typename": "StackConfigVendorTerraform"}}
	module := node{"integrations": node{}}
	// root and legacy must be skipped.
	spaces := []node{
		{"id": "root", "name": "root", "parentSpace": nil},
		{"id": "legacy", "name": "legacy", "parentSpace": "root"},
		{"id": "a", "name": "a", "parentSpace": "root"},
		{"id": "b", "name": "b", "parentSpace": "a"},
	}

	for _, tc := range []struct {
		typeName, field, identity string
		nodes                     []node
		want                      []string
	}{
		{"spacelift_stack", "searchStacks", "ulid", ulids(stack), []string{"U1", "U2"}},
		{"spacelift_module", "searchModules", "ulid", ulids(module), []string{"U1", "U2"}},
		{"spacelift_context", "searchContexts", "ulid", ulids(nil), []string{"U1", "U2"}},
		{"spacelift_policy", "searchPolicies", "ulid", ulids(nil), []string{"U1", "U2"}},
		{"spacelift_worker_pool", "searchWorkerPools", "id", ids(), []string{"A", "B"}},
		{"spacelift_aws_integration", "awsIntegrations", "id", ids(), []string{"A", "B"}},
		{"spacelift_azure_integration", "azureIntegrations", "id", ids(), []string{"A", "B"}},
		{"spacelift_space", "searchSpaces", "id", spaces, []string{"a", "b"}},
	} {
		t.Run(tc.typeName, func(t *testing.T) {
			url, _ := fakeListAPI(t, tc.field, tc.nodes)
			server, schemas := newListTestProvider(t, url)
			resourceType := schemas.ResourceSchemas[tc.typeName].ValueType()
			identityType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{tc.identity: tftypes.String}}

			var got []string
			for _, result := range listResults(t, server, schemas, tc.typeName, nil, 0, true) {
				identity := stringAttribute(t, result.Identity.IdentityData, identityType, tc.identity)
				if attr := stringAttribute(t, result.Resource, resourceType, tc.identity); attr != identity {
					t.Errorf("resource %s %q does not match identity %q", tc.identity, attr, identity)
				}
				if result.DisplayName == "" {
					t.Errorf("%s has no display name", identity)
				}
				got = append(got, identity)
			}
			assertStrings(t, got, tc.want...)

			limited := listResults(t, server, schemas, tc.typeName, nil, 1, false)
			if len(limited) != 1 || limited[0].Resource != nil {
				t.Errorf("limit 1 without resource: got %+v", limited)
			}
		})
	}
}

// TestListPredicates checks the filters reach the search API as predicates.
func TestListPredicates(t *testing.T) {
	str := func(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
	config := func(key string, value tftypes.Value) map[string]tftypes.Value {
		return map[string]tftypes.Value{
			"name":   str("prod-*"),
			"labels": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{str("x"), str("y")}),
			key:      value,
		}
	}

	const (
		name   = `{"field":"name","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":["prod-*"]}}`
		labels = `{"field":"label","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":null,"stringEquals":["x"]}},` +
			`{"field":"label","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":null,"stringEquals":["y"]}}`
	)
	space := func(path string, descendants bool) string {
		return fmt.Sprintf(`{"field":"space","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":null,"hierarchyNode":{"nodes":[{"equals":%q,"includeDescendants":%v}]}}}`, path, descendants)
	}

	for _, tc := range []struct {
		typeName, field string
		config          map[string]tftypes.Value
		want            string
	}{
		{"spacelift_stack", "searchStacks", config("space_id", str("a")),
			"[" + name + "," + space("root/a", false) + "," + labels + "]"},
		{"spacelift_policy", "searchPolicies", config("type", str("PLAN")),
			"[" + name + "," + labels + `,{"field":"type","constraint":{"booleanEquals":null,"enumEquals":["PLAN"],"stringMatches":null}}]`},
		// Spaces always search the parent's whole subtree.
		{"spacelift_space", "searchSpaces", config("parent_space_id", str("a")),
			"[" + name + "," + space("root/a", true) + "," + labels + "]"},
	} {
		t.Run(tc.typeName, func(t *testing.T) {
			url, predicates := fakeListAPI(t, tc.field, nil)
			server, schemas := newListTestProvider(t, url)
			listResults(t, server, schemas, tc.typeName, tc.config, 0, false)

			if len(*predicates) != 1 || (*predicates)[0] != tc.want {
				t.Errorf("predicates:\n got %v\nwant %s", *predicates, tc.want)
			}
		})
	}
}

// TestListConfigErrors checks invalid configs fail before calling the API.
func TestListConfigErrors(t *testing.T) {
	for _, tc := range []struct {
		typeName string
		config   map[string]tftypes.Value
		want     string
	}{
		{"spacelift_policy", map[string]tftypes.Value{"type": tftypes.NewValue(tftypes.String, "PLANN")}, "invalid policy type"},
		{"spacelift_stack", map[string]tftypes.Value{"include_child_spaces": tftypes.NewValue(tftypes.Bool, true)}, "missing space_id"},
		{"spacelift_space", map[string]tftypes.Value{"include_child_spaces": tftypes.NewValue(tftypes.Bool, true)}, "missing parent_space_id"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("unexpected API call") }))
			t.Cleanup(server.Close)
			provider, schemas := newListTestProvider(t, server.URL)

			stream, err := provider.ListResource(t.Context(), &tfprotov6.ListResourceRequest{
				TypeName: tc.typeName,
				Config:   dynamicValue(t, schemas.ListResourceSchemas[tc.typeName].ValueType(), tc.config),
			})
			if err != nil {
				t.Fatal(err)
			}

			var summaries []string
			for result := range stream.Results {
				for _, d := range result.Diagnostics {
					summaries = append(summaries, d.Summary)
				}
			}
			assertStrings(t, summaries, tc.want)
		})
	}
}

// fakeListAPI serves nodes as the output of the query field, and the space
// tree root > a. It returns its URL and the predicates of each search request.
func fakeListAPI(t *testing.T, field string, nodes []node) (string, *[]string) {
	var predicates []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables struct {
				Input struct {
					After      *string         `json:"after"`
					Predicates json.RawMessage `json:"predicates"`
				} `json:"input"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("could not decode request: %v", err)
		}

		var output any
		switch {
		case r.Header.Get("Spacelift-GraphQL-Query") == "ListSpaceTree":
			output = node{"spaces": []node{{"id": "root", "parentSpace": nil}, {"id": "a", "parentSpace": "root"}}}
		case !strings.HasPrefix(field, "search"):
			output = node{field: nodes}
		default:
			predicates = append(predicates, string(body.Variables.Input.Predicates))
			// The last node is on the second page.
			page, hasNext := nodes[:max(len(nodes)-1, 0)], len(nodes) > 0
			if body.Variables.Input.After != nil {
				page, hasNext = nodes[len(nodes)-1:], false
			}
			edges := []node{}
			for _, n := range page {
				edges = append(edges, node{"node": n})
			}
			output = node{field: node{"edges": edges, "pageInfo": node{"hasNextPage": hasNext, "endCursor": "c1"}}}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(node{"data": output}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	return server.URL, &predicates
}

func listResults(t *testing.T, server tfprotov6.ProviderServerWithListResource, schemas *tfprotov6.GetProviderSchemaResponse, typeName string, config map[string]tftypes.Value, limit int64, includeResource bool) []tfprotov6.ListResourceResult {
	t.Helper()

	stream, err := server.ListResource(t.Context(), &tfprotov6.ListResourceRequest{
		TypeName:        typeName,
		Config:          dynamicValue(t, schemas.ListResourceSchemas[typeName].ValueType(), config),
		Limit:           limit,
		IncludeResource: includeResource,
	})
	if err != nil {
		t.Fatal(err)
	}

	var results []tfprotov6.ListResourceResult
	for result := range stream.Results {
		for _, d := range result.Diagnostics {
			t.Fatalf("list: %s: %s", d.Summary, d.Detail)
		}
		results = append(results, result)
	}
	return results
}

// newListTestProvider returns the muxed provider, configured to talk to the
// fake API at url, and its schemas.
func newListTestProvider(t *testing.T, url string) (tfprotov6.ProviderServerWithListResource, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"aud": url}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SPACELIFT_API_TOKEN", token)
	// A full API key takes precedence over the token, and CI sets one.
	for _, name := range []string{"SPACELIFT_API_KEY_ENDPOINT", "SPACELIFT_API_KEY_ID", "SPACELIFT_API_KEY_SECRET"} {
		t.Setenv(name, "")
	}

	muxed, err := testAccProtoV6MuxProviderFactories()["spacelift"]()
	if err != nil {
		t.Fatal(err)
	}
	server, ok := muxed.(tfprotov6.ProviderServerWithListResource)
	if !ok {
		t.Fatal("the muxed provider does not serve list resources")
	}

	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}

	configure, err := server.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
		Config: dynamicValue(t, schemas.Provider.ValueType(), nil),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range configure.Diagnostics {
		t.Fatalf("configure: %s: %s", d.Summary, d.Detail)
	}

	return server, schemas
}

// dynamicValue builds an object of typ, with the given attributes set and the
// rest null.
func dynamicValue(t *testing.T, typ tftypes.Type, set map[string]tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()
	attrs := map[string]tftypes.Value{}
	for name, attrType := range typ.(tftypes.Object).AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range set {
		attrs[name] = value
	}
	value, err := tfprotov6.NewDynamicValue(typ, tftypes.NewValue(typ, attrs))
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func stringAttribute(t *testing.T, value *tfprotov6.DynamicValue, typ tftypes.Type, name string) string {
	t.Helper()

	object, err := value.Unmarshal(typ)
	if err != nil {
		t.Fatal(err)
	}
	var attrs map[string]tftypes.Value
	if err := object.As(&attrs); err != nil {
		t.Fatal(err)
	}
	var out string
	if err := attrs[name].As(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func assertStrings(t *testing.T, got []string, want ...string) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
