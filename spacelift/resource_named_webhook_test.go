package spacelift

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestNamedWebhookResource(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_named_webhook.test"

	t.Run("attach a webhook to root space with all fields filled", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint         = "%s"
					space_id         = "root"
					name             = "testing-named-%s"
					labels           = ["1", "2"]
					secret           = "super-secret"
					enabled          = true
					retry_on_failure = false
				}
			`, endpoint, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					SetEquals("labels", "1", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("attach a webhook to root space with default values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint         = "%s"
					space_id         = "root"
					name             = "testing-named-%s"
					enabled          = true
					retry_on_failure = false
				}
			`, endpoint, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("secret", Equals("")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					AttributeNotPresent("labels"),
				),
			},
		})
	})

	t.Run("attach a webhook to root space with write-only values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_named_webhook" "test" {
					endpoint          = "%s"
					space_id          = "root"
					name              = "testing-named-%s"
					enabled           = true
					secret_wo         = "super-secret"
					secret_wo_version = 1
					retry_on_failure  = false
				}
			`, endpoint, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.net"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.net")),
					Attribute("space_id", Equals("root")),
					Attribute("name", Equals(fmt.Sprintf("testing-named-%s", randomID))),
					Attribute("enabled", Equals("true")),
					AttributeNotPresent("labels"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version"},
			},
		})
	})
}

func TestNamedWebhookSendsRetryOnFailureFalse(t *testing.T) {
	t.Parallel()

	var createRetryOnFailure, updateRetryOnFailure *bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Query     string `json:"query"`
			Variables struct {
				Input struct {
					RetryOnFailure *bool `json:"retryOnFailure"`
				} `json:"input"`
			} `json:"variables"`
		}
		_ = json.Unmarshal(body, &req)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(req.Query, "namedWebhooksIntegrationCreate"):
			createRetryOnFailure = req.Variables.Input.RetryOnFailure
			_, _ = w.Write([]byte(`{"data":{"namedWebhooksIntegrationCreate":{"id":"nwh-1","enabled":true}}}`))
		case strings.Contains(req.Query, "namedWebhooksIntegrationUpdate"):
			updateRetryOnFailure = req.Variables.Input.RetryOnFailure
			_, _ = w.Write([]byte(`{"data":{"namedWebhooksIntegrationUpdate":{"id":"nwh-1"}}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"namedWebhooksIntegration":{"id":"nwh-1","name":"wh","endpoint":"https://example.com/hook","enabled":true,"space":{"id":"root"},"labels":[],"retryOnFailure":false}}}`))
		}
	}))
	t.Cleanup(server.Close)

	client := internal.NewClient(server.URL, "token", nil, nil)

	d := resourceNamedWebhook().Data(nil)
	if err := d.Set("enabled", true); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("endpoint", "https://example.com/hook"); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("space_id", "root"); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("name", "wh"); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("retry_on_failure", false); err != nil {
		t.Fatal(err)
	}

	if diags := resourceNamedWebhookCreate(context.Background(), d, client); diags.HasError() {
		t.Fatalf("create failed: %v", diags)
	}
	if createRetryOnFailure == nil || *createRetryOnFailure {
		t.Fatalf("create retryOnFailure = %v, want false", createRetryOnFailure)
	}

	if diags := resourceNamedWebhookUpdate(context.Background(), d, client); diags.HasError() {
		t.Fatalf("update failed: %v", diags)
	}
	if updateRetryOnFailure == nil || *updateRetryOnFailure {
		t.Fatalf("update retryOnFailure = %v, want false", updateRetryOnFailure)
	}
}
