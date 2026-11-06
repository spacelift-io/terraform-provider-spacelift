package spacelift

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func TestAzureIntegrationConsentLifecycle(t *testing.T) {
	t.Parallel()

	responses := []string{
		`{"data":{"azureIntegrationConsentConfirm":{"adminConsentProvided":true}}}`,
		`{"data":{"azureIntegration":{"adminConsentProvided":true}}}`,
		`{"data":{"azureIntegrationConsentClear":{"adminConsentProvided":false}}}`,
	}
	client := azureIntegrationConsentTestClient(t, responses,
		[]string{"azureIntegrationConsentConfirm", "azureIntegration", "azureIntegrationConsentClear"})
	r := resourceAzureIntegrationConsent()
	d := r.Data(nil)
	if err := d.Set("integration_id", "integration-id"); err != nil {
		t.Fatal(err)
	}
	if diags := r.CreateContext(t.Context(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	if d.Id() != "integration-id" || d.Get("admin_consent_provided") != true {
		t.Fatalf("unexpected state after create: %#v", d.State())
	}

	imported := r.Data(nil)
	imported.SetId(d.Id())
	states, err := r.Importer.StateContext(t.Context(), imported, client)
	if err != nil || len(states) != 1 {
		t.Fatalf("import: states=%v, err=%v", states, err)
	}
	if diags := r.ReadContext(t.Context(), states[0], client); diags.HasError() {
		t.Fatal(diags)
	}
	if states[0].Get("integration_id") != "integration-id" || states[0].Get("admin_consent_provided") != true {
		t.Fatalf("unexpected imported state: %#v", states[0].State())
	}

	if diags := r.DeleteContext(t.Context(), d, client); diags.HasError() {
		t.Fatal(diags)
	}
	if d.Id() != "" {
		t.Fatalf("resource still in state after delete: %q", d.Id())
	}
}

func TestAzureIntegrationConsentFailuresAndDrift(t *testing.T) {
	t.Parallel()

	const apiError = `{"errors":[{"message":"access denied"}]}`
	cases := []struct {
		name      string
		operation func(context.Context, *schema.ResourceData, any) diag.Diagnostics
		initialID string
		responses []string
		fields    []string
		wantError bool
		wantID    string
	}{
		{
			name: "confirmation rejected", operation: resourceAzureIntegrationConsentCreate,
			responses: []string{apiError}, fields: []string{"azureIntegrationConsentConfirm"}, wantError: true,
		},
		{
			name: "confirmation not recorded", operation: resourceAzureIntegrationConsentCreate,
			responses: []string{`{"data":{"azureIntegrationConsentConfirm":{"adminConsentProvided":false}}}`},
			fields:    []string{"azureIntegrationConsentConfirm"}, wantError: true,
		},
		{
			name: "missing integration", operation: resourceAzureIntegrationConsentRead, initialID: "integration-id",
			responses: []string{`{"data":{"azureIntegration":null}}`}, fields: []string{"azureIntegration"},
		},
		{
			name: "confirmation cleared externally", operation: resourceAzureIntegrationConsentRead, initialID: "integration-id",
			responses: []string{`{"data":{"azureIntegration":{"adminConsentProvided":false}}}`},
			fields:    []string{"azureIntegration"},
		},
		{
			name: "refresh keeps confirmation", operation: resourceAzureIntegrationConsentRead, initialID: "integration-id",
			responses: []string{`{"data":{"azureIntegration":{"adminConsentProvided":true}}}`},
			fields:    []string{"azureIntegration"}, wantID: "integration-id",
		},
		{
			name: "refresh error retains state", operation: resourceAzureIntegrationConsentRead, initialID: "integration-id",
			responses: []string{apiError}, fields: []string{"azureIntegration"}, wantError: true, wantID: "integration-id",
		},
		{
			name: "clear failure retains state", operation: resourceAzureIntegrationConsentDelete, initialID: "integration-id",
			responses: []string{apiError}, fields: []string{"azureIntegrationConsentClear"}, wantError: true, wantID: "integration-id",
		},
		{
			name: "clear not recorded", operation: resourceAzureIntegrationConsentDelete, initialID: "integration-id",
			responses: []string{`{"data":{"azureIntegrationConsentClear":{"adminConsentProvided":true}}}`},
			fields:    []string{"azureIntegrationConsentClear"}, wantError: true, wantID: "integration-id",
		},
		{
			name: "delete missing integration", operation: resourceAzureIntegrationConsentDelete, initialID: "integration-id",
			responses: []string{`{"data":{"azureIntegrationConsentClear":null}}`}, fields: []string{"azureIntegrationConsentClear"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := azureIntegrationConsentTestClient(t, tc.responses, tc.fields)
			d := resourceAzureIntegrationConsent().Data(nil)
			d.SetId(tc.initialID)
			if err := d.Set("integration_id", "integration-id"); err != nil {
				t.Fatal(err)
			}
			diags := tc.operation(t.Context(), d, client)
			if diags.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if d.Id() != tc.wantID {
				t.Fatalf("got ID %q, want %q", d.Id(), tc.wantID)
			}
			if !tc.wantError && d.Id() != "" && d.Get("admin_consent_provided") != true {
				t.Fatalf("confirmation missing after refresh: %v", d.Get("admin_consent_provided"))
			}
		})
	}
}

func azureIntegrationConsentTestClient(t *testing.T, responses, fields []string) *internal.Client {
	t.Helper()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		index := int(requests.Add(1)) - 1
		if index >= len(responses) {
			t.Errorf("unexpected request %d", index)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var request struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/graphql" || r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("unexpected API request: %s, authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		if !strings.Contains(strings.ReplaceAll(request.Query, " ", ""), fields[index]+"(id:$id)") || request.Variables["id"] != "integration-id" {
			t.Errorf("unexpected GraphQL request: %+v", request)
		}
		if strings.Contains(request.Query, "servicePrincipalObjectId") {
			t.Error("consent resource requested the service principal object ID")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responses[index]))
	}))
	t.Cleanup(func() {
		server.Close()
		if got := int(requests.Load()); got != len(responses) {
			t.Errorf("got %d API requests, want %d", got, len(responses))
		}
	})
	return internal.NewClient(server.URL, "token", nil, nil)
}
