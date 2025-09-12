package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleData(t *testing.T) {
	t.Run("basic test", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
			resource "spacelift_module" "test" {
                name            = "test-module-%s"
				administrative  = true
				branch          = "master"
				description     = "description"
				labels          = ["one", "two"]
				git_sparse_checkout_paths = ["module"]
				repository      = "terraform-bacon-tasty"
				shared_accounts = ["spacelift-io"]
			}
			data "spacelift_module" "test" {
				module_id = spacelift_module.test.id
			}
		`, randomID),
			Check: Resource(
				"data.spacelift_module.test",
				Attribute("id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
				Attribute("administrative", Equals("true")),
				Attribute("branch", Equals("master")),
				Attribute("description", Equals("description")),
				SetEquals("labels", "one", "two"),
				Attribute("name", Equals(fmt.Sprintf("test-module-%s", randomID))),
				Attribute("project_root", Equals("")),
				SetEquals("git_sparse_checkout_paths", "module"),
				Attribute("repository", Equals("terraform-bacon-tasty")),
				SetEquals("shared_accounts", "spacelift-io"),
				Attribute("space_shares.#", Equals("0")),
				Attribute("terraform_provider", Equals("default")),
			),
		}})
	})

	t.Run("with terraform_workflow_tool defaulted", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name            = "test-module-%s"
					administrative  = true
					branch          = "master"
					repository      = "terraform-bacon-tasty"
				}
				data "spacelift_module" "test" {
					module_id = spacelift_module.test.id
				}
			`, randomID),
				Check: Resource(
					"data.spacelift_module.test",
					Attribute("workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			},
		})
	})

	t.Run("with terraform_workflow_tool set", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name            = "test-module-%s"
					administrative  = true
					branch          = "master"
					repository      = "terraform-bacon-tasty"
					workflow_tool   = "CUSTOM"
				}
				data "spacelift_module" "test" {
					module_id = spacelift_module.test.id
				}
			`, randomID),
				Check: Resource(
					"data.spacelift_module.test",
					Attribute("workflow_tool", Equals("CUSTOM")),
				),
			},
		})
	})

	t.Run("with Raw Git", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name            = "test-module-%s"
					administrative  = false
					branch          = "main"
					repository      = "terraform-bacon-tasty"

					raw_git {
						namespace = "bacon"
						url       = "https://gist.github.com/d8d18c7c2841b578de22be34cb5943f5.git"
					}
				}
				data "spacelift_module" "test" {
					module_id = spacelift_module.test.id
				}
			`, randomID),
				Check: Resource(
					"data.spacelift_module.test",
					Nested("raw_git",
						CheckInList(
							Attribute("namespace", Equals("bacon")),
							Attribute("url", Equals("https://gist.github.com/d8d18c7c2841b578de22be34cb5943f5.git")),
						),
					),
				),
			},
		})
	})
}

func TestModuleDataSpace(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_module" "test" {
                name            = "test-module-%s"
				administrative  = true
				branch          = "master"
				description     = "description"
				labels          = ["one", "two"]
				repository      = "terraform-bacon-tasty"
				shared_accounts = ["spacelift-io"]
				space_id        = "root"
			}

			data "spacelift_module" "test" {
				module_id = spacelift_module.test.id
			}
		`, randomID),
		Check: Resource(
			"data.spacelift_module.test",
			Attribute("id", Equals(fmt.Sprintf("terraform-default-test-module-%s", randomID))),
			Attribute("administrative", Equals("true")),
			Attribute("branch", Equals("master")),
			Attribute("description", Equals("description")),
			SetEquals("labels", "one", "two"),
			Attribute("name", Equals(fmt.Sprintf("test-module-%s", randomID))),
			Attribute("project_root", Equals("")),
			Attribute("repository", Equals("terraform-bacon-tasty")),
			Attribute("space_id", Equals("root")),
			SetEquals("shared_accounts", "spacelift-io"),
			Attribute("space_shares.#", Equals("0")),
			Attribute("terraform_provider", Equals("default")),
		),
	}})
}

func TestModuleDataSpaceShares(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	spaces := `
		resource "spacelift_space" "test_space_1" {
			name = "test-space-1"
		}
		resource "spacelift_space" "test_space_2" {
			name = "test-space-2"
		}`

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			%s
			resource "spacelift_module" "test" {
				name            = "test-module-%s"
				branch          = "master"
				repository      = "terraform-bacon-tasty"
				space_shares    = [spacelift_space.test_space_1.id, spacelift_space.test_space_2.id]
			}

			data "spacelift_module" "test" {
				module_id = spacelift_module.test.id
			}
		`, spaces, randomID),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.spacelift_module.test", "id", fmt.Sprintf("terraform-default-test-module-%s", randomID)),
			// Check space_shares contains exactly 2 elements
			resource.TestCheckResourceAttr("data.spacelift_module.test", "space_shares.#", "2"),
			// Verify the space IDs are in the set (order doesn't matter)
			resource.TestCheckTypeSetElemAttrPair("data.spacelift_module.test", "space_shares.*", "spacelift_space.test_space_1", "id"),
			resource.TestCheckTypeSetElemAttrPair("data.spacelift_module.test", "space_shares.*", "spacelift_space.test_space_2", "id"),
		),
	}})
}
