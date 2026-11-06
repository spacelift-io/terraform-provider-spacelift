package spacelift

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWebhookCreateHonorsEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		requested   bool
		returned    bool
		wantErr     string
		wantID      string
		wantEnabled string
	}{
		{
			name:        "disabled webhook is created",
			requested:   false,
			returned:    false,
			wantID:      "wh-disabled",
			wantEnabled: "false",
		},
		{
			name:      "requested enabled webhook that comes back disabled is rejected",
			requested: true,
			returned:  false,
			wantErr:   "webhook not activated",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				query := string(body)
				w.Header().Set("Content-Type", "application/json")
				switch {
				case strings.Contains(query, "webhooksIntegrationCreate"):
					fmt.Fprintf(w, `{"data":{"webhooksIntegrationCreate":{"id":"wh-disabled","enabled":%t}}}`, tc.returned)
				default:
					_, _ = w.Write([]byte(`{"data":{"stack":{"id":"stack-1","integrations":{"webhooks":[{"id":"wh-disabled","enabled":false,"endpoint":"https://example.com/hook","retryOnFailure":false}]}}}}`))
				}
			}))
			t.Cleanup(server.Close)

			d := resourceWebhook().Data(nil)
			if err := d.Set("enabled", tc.requested); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("endpoint", "https://example.com/hook"); err != nil {
				t.Fatal(err)
			}
			if err := d.Set("stack_id", "stack-1"); err != nil {
				t.Fatal(err)
			}

			diags := resourceWebhookCreate(context.Background(), d, internal.NewClient(server.URL, "token", nil, nil))
			if tc.wantErr != "" {
				if !diags.HasError() {
					t.Fatal("expected an error")
				}
				if !strings.Contains(diags[0].Summary, tc.wantErr) {
					t.Fatalf("got %q, want it to contain %q", diags[0].Summary, tc.wantErr)
				}
				return
			}
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if d.Id() != tc.wantID {
				t.Fatalf("id = %q, want %q", d.Id(), tc.wantID)
			}
			if got := fmt.Sprint(d.Get("enabled")); got != tc.wantEnabled {
				t.Fatalf("enabled = %s, want %s", got, tc.wantEnabled)
			}
		})
	}
}

func TestWebhookResource(t *testing.T) {
	t.Parallel()

	const resourceName = "spacelift_webhook.test"

	t.Run("with a stack", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_webhook" "test" {
					stack_id         = spacelift_stack.test.id
					endpoint         = "%s"
					secret           = "very-very-secret"
					retry_on_failure = false
				}
			`, randomID, endpoint)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.org"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
					Attribute("secret", Equals("")),
				),
			},
			{
				ResourceName:        resourceName,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("stack/test-stack-%s/", randomID),
				ImportStateVerify:   true,
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})

	t.Run("with a module", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_module" "test" {
                name       = "test-module-%s"
				branch     = "master"
				repository = "terraform-bacon-tasty"
			}

			resource "spacelift_webhook" "test" {
				module_id        = spacelift_module.test.id
				endpoint         = "https://bacon.org"
				secret           = "very-very-secret"
				retry_on_failure = false
			}
		`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
					Attribute("secret", Equals("")),
				),
			},
			{
				ResourceName:        resourceName,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("module/terraform-default-test-module-%s/", randomID),
				ImportStateVerify:   true,
			},
		})
	})

	t.Run("with a stack and write-only values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(endpoint string) string {
			return fmt.Sprintf(`
				resource "spacelift_stack" "test" {
					branch     = "master"
					repository = "demo"
					name       = "Test stack %s"
				}

				resource "spacelift_webhook" "test" {
					stack_id          = spacelift_stack.test.id
					endpoint          = "%s"
					secret_wo         = "very-very-secret"
					secret_wo_version = 1
					retry_on_failure  = false
				}
			`, randomID, endpoint)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("https://bacon.org"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("endpoint", Equals("https://bacon.org")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdPrefix:     fmt.Sprintf("stack/test-stack-%s/", randomID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "secret_wo_version"},
			},
			{
				Config: config("https://cabbage.org"),
				Check:  Resource(resourceName, Attribute("endpoint", Equals("https://cabbage.org"))),
			},
		})
	})
}
