package spacelift

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestStackDestructorAndActivatorImport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}

		w.Header().Set("Content-Type", "application/json")
		if request.Variables["id"] != "my-stack" {
			_, _ = w.Write([]byte(`{"data":{"stack":null}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"stack":{"id":"my-stack","isDisabled":true}}}`))
	}))
	t.Cleanup(server.Close)

	claims := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"aud":[%q]}`, server.URL)))
	t.Setenv("SPACELIFT_API_TOKEN", "eyJhbGciOiJub25lIn0."+claims+".")

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "spacelift_stack_destructor" "test" {
						stack_id = "my-stack"
					}
				`,
				ResourceName:  "spacelift_stack_destructor.test",
				ImportState:   true,
				ImportStateId: "my-stack",
				ImportStateCheck: importedAttributes(map[string]string{
					"stack_id":     "my-stack",
					"discard_runs": "false",
				}),
			},
			{
				Config: `
					resource "spacelift_stack_activator" "test" {
						stack_id = "my-stack"
						enabled  = false
					}
				`,
				ResourceName:  "spacelift_stack_activator.test",
				ImportState:   true,
				ImportStateId: "my-stack",
				ImportStateCheck: importedAttributes(map[string]string{
					"stack_id": "my-stack",
					"enabled":  "false",
				}),
			},
		},
	})
}

func importedAttributes(want map[string]string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("expected 1 imported resource, got %d", len(states))
		}
		for key, value := range want {
			if got := states[0].Attributes[key]; got != value {
				return fmt.Errorf("%s: got %q, want %q", key, got, value)
			}
		}
		return nil
	}
}
