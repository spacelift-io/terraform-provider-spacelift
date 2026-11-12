package spacelift

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

// Runs the create of `spacelift_named_webhook_secret_header` against a stub
// Spacelift API and returns the header entry the provider sent.
//
// Write-only arguments are only readable from the raw configuration, so the
// configuration is handed over the same way the plugin server hands it to a
// resource: shimmed into a diff that still carries the original cty value.
func namedWebhookSecretHeaderCreateEntry(t *testing.T, config map[string]cty.Value) structs.KeyValuePair {
	t.Helper()

	var mutationVariables map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := `{"data":{"namedWebhooksIntegration":{"id":"webhook-id","secretHeaders":["X-Token"]}}}`
		if r.Header.Get("Spacelift-GraphQL-Mutation") == "NamedWebhookSecretHeaderSet" {
			mutationVariables = request.Variables
			response = `{"data":{"namedWebhooksIntegrationSetHeaders":{"id":"webhook-id"}}}`
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	r := resourceNamedWebhookSecretHeader()
	configSchema := r.CoreConfigSchema()

	attributes := map[string]cty.Value{}
	for name, attribute := range configSchema.Attributes {
		value, ok := config[name]
		if !ok {
			value = cty.NullVal(attribute.Type)
		}
		attributes[name] = value
	}
	configValue := cty.ObjectVal(attributes)

	instanceDiff, err := r.Diff(context.Background(), &terraform.InstanceState{}, terraform.NewResourceConfigShimmed(configValue, configSchema), nil)
	if err != nil {
		t.Fatalf("could not plan the resource: %v", err)
	}
	instanceDiff.RawConfig = configValue

	_, diags := r.Apply(context.Background(), &terraform.InstanceState{}, instanceDiff, internal.NewClient(server.URL, "token", nil, nil))
	if diags.HasError() {
		t.Fatalf("create returned errors: %v", diags)
	}

	if mutationVariables == nil {
		t.Fatal("no mutation was sent")
	}
	input, ok := mutationVariables["input"].(map[string]any)
	if !ok {
		t.Fatalf("mutation has no input: %v", mutationVariables)
	}
	entries, ok := input["entries"].([]any)
	if !ok || len(entries) != 1 {
		t.Fatalf("mutation input has no single entry: %v", input)
	}

	var entry structs.KeyValuePair
	body, err := json.Marshal(entries[0])
	if err != nil {
		t.Fatalf("could not marshal entry: %v", err)
	}
	if err := json.Unmarshal(body, &entry); err != nil {
		t.Fatalf("could not unmarshal entry: %v", err)
	}

	return entry
}

func TestNamedWebhookSecretHeaderValue(t *testing.T) {
	t.Parallel()

	// `value_wo` is only picked up when the configuration also has a
	// `value_wo_version`, so both names have to be looked up exactly as they are
	// spelled in the resource schema.
	t.Run("write-only value", func(t *testing.T) {
		entry := namedWebhookSecretHeaderCreateEntry(t, map[string]cty.Value{
			"webhook_id":       cty.StringVal("webhook-id"),
			"key":              cty.StringVal("X-Token"),
			"value_wo":         cty.StringVal("write-only-secret"),
			"value_wo_version": cty.StringVal("1"),
		})

		if entry.Key != "X-Token" {
			t.Errorf("expected key %q, got %q", "X-Token", entry.Key)
		}
		if entry.Value != "write-only-secret" {
			t.Errorf("expected the write-only value to be sent, got %q", entry.Value)
		}
	})

	t.Run("plain value", func(t *testing.T) {
		entry := namedWebhookSecretHeaderCreateEntry(t, map[string]cty.Value{
			"webhook_id": cty.StringVal("webhook-id"),
			"key":        cty.StringVal("X-Token"),
			"value":      cty.StringVal("plaintext"),
		})

		if entry.Value != "plaintext" {
			t.Errorf("expected %q, got %q", "plaintext", entry.Value)
		}
	})
}

func TestNamedWebhookSecretHeaderResource(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_named_webhook_secret_header.test-secret"

	t.Run("attach a webhook to root space with all fields filled", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint = "https://bacon.org"
					space_id = "root"
					name     = "%s"
					labels   = ["1", "2"]
					secret   = "super-secret"
					enabled  = true
				}

				resource "spacelift_named_webhook_secret_header" "test-secret" {
					webhook_id = spacelift_named_webhook.test.id
					key        = "thisisakey"
					value      = "thisisavalue"
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(fmt.Sprintf("%s/%s", randomID, "thisisakey"))),
					Attribute("key", Equals("thisisakey")),
					Attribute("value", Equals("thisisavalue")),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
		})
	})

	t.Run("attach a webhook to root space with all write-only fields", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint          = "https://bacon.org"
					space_id          = "root"
					name              = "%s"
					labels            = ["1", "2"]
					secret_wo         = "super-secret"
					secret_wo_version = 1
					enabled           = true
				}

				resource "spacelift_named_webhook_secret_header" "test-secret" {
					webhook_id            = spacelift_named_webhook.test.id
					key                   = "thisisakey"
					value_wo              = "thisisavalue"
					value_wo_version      = 1
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(fmt.Sprintf("%s/%s", randomID, "thisisakey"))),
					Attribute("key", Equals("thisisakey")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version", "value", "value_wo_version"},
			},
		})
	})
}
