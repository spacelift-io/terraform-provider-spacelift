package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestModuleResource(t *testing.T) {
	t.Run("with GitHub", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string, protectFromDeletion bool, localPreview bool) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name                  = "github-module-%s"
					administrative        = true
					branch                = "master"
					description           = "%s"
					labels                = ["one", "two"]
					git_sparse_checkout_paths = ["module"]
					enable_local_preview  = %t
					protect_from_deletion = %t
					repository            = "terraform-bacon-tasty"
					shared_accounts       = ["foo-subdomain", "bar-subdomain"]
				}
			`, randomID, description, localPreview, protectFromDeletion)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description", true, false),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("terraform-default-github-module-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals(fmt.Sprintf("github-module-%s", randomID))),
					AttributeNotPresent("project_root"),
					SetEquals("git_sparse_checkout_paths", "module"),
					Attribute("enable_local_preview", Equals("false")),
					Attribute("protect_from_deletion", Equals("true")),
					Attribute("public", Equals("false")),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("default")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description", false, true),
				Check: Resource(
					"spacelift_module.test",
					Attribute("description", Equals("new description")),
					Attribute("enable_local_preview", Equals("true")),
					Attribute("protect_from_deletion", Equals("false")),
					Attribute("public", Equals("false")),
				),
			},
		})
	})

	t.Run("with Raw Git", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name           = "raw-git-%s"
					administrative = false
					branch         = "main"
					repository     = "terraform-bacon-tasty"

					raw_git {
						namespace = "bacon"
						url       = "https://gist.github.com/d8d18c7c2841b578de22be34cb5943f5.git"
					}
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					"spacelift_module.test",
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

	t.Run("with public", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func() string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name           = "public-test-%s"
					administrative = false
					branch         = "main"
					repository     = "terraform-bacon-tasty"
					public         = true

					raw_git {
						namespace = "bacon"
						url       = "https://gist.github.com/d8d18c7c2841b578de22be34cb5943f5.git"
					}
				}
			`, randomID)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(),
				Check: Resource(
					"spacelift_module.test",
					Attribute("public", Equals("true")),
				),
			},
		})
	})

	t.Run("project root and custom name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(projectRoot string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
                    name               = "project-root-%s"
					administrative     = true
					branch             = "master"
					description        = "description"
					labels             = ["one", "two"]
                    project_root       = "%s"
					repository         = "terraform-bacon-tasty"
					shared_accounts    = ["foo-subdomain", "bar-subdomain"]
                    terraform_provider = "papaya"
				}
			`, randomID, projectRoot)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("test-root/ab"),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("terraform-papaya-project-root-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals(fmt.Sprintf("project-root-%s", randomID))),
					Attribute("project_root", Equals("test-root/ab")),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("papaya")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("test-root/bc"),
				Check:  Resource("spacelift_module.test", Attribute("project_root", Equals("test-root/bc"))),
			},
		})
	})

	for _, name := range []string{
		"github-Module",
		"github-module-",
		"_github-module",
		"0github-module",
	} {
		t.Run("invalid name", func(t *testing.T) {
			testSteps(t, []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						resource "spacelift_module" "test" {
							name                  = "%s"
							branch                = "master"
							repository            = "terraform-bacon-tasty"
						}
			`, name),
					ExpectError: regexp.MustCompile("must start and end with lowercase letter and may only contain lowercase letters, digits, dashes and underscores"),
				},
			})
		})
	}

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_module" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = ["one", "two"]
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_module.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_module" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = []
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_module.test",
					SetEquals("labels"),
				),
			},
		})
	})

	t.Run("with workflow_tool", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			// Defaults to TERRAFORM_FOSS
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "workflow_tool" {
                    name               = "workflow-tool-%s"
					branch             = "master"
					repository         = "terraform-bacon-tasty"
                    terraform_provider = "papaya"
				}
			`, randomID),
				Check: Resource(
					"spacelift_module.workflow_tool",
					Attribute("workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			},
			// Can update the tool
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "workflow_tool" {
                    name               = "workflow-tool-%s"
					branch             = "master"
					repository         = "terraform-bacon-tasty"
                    terraform_provider = "papaya"
					workflow_tool      = "CUSTOM"
				}
			`, randomID),
				Check: Resource(
					"spacelift_module.workflow_tool",
					Attribute("workflow_tool", Equals("CUSTOM")),
				),
			},
			// Can create a module with OPEN_TOFU
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "workflow_tool_open_tofu" {
			        name               = "workflow-tool-open-tofu-%s"
					branch             = "master"
					repository         = "terraform-bacon-tasty"
			        terraform_provider = "papaya"
					workflow_tool      = "OPEN_TOFU"
				}
			`, randomID),
				Check: Resource(
					"spacelift_module.workflow_tool_open_tofu",
					Attribute("workflow_tool", Equals("OPEN_TOFU")),
				),
			},
			// Can create a module with TERRAFORM_FOSS
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "workflow_tool_terraform_foss" {
			        name               = "workflow-tool-terraform-foss-%s"
					branch             = "master"
					repository         = "terraform-bacon-tasty"
			        terraform_provider = "papaya"
					workflow_tool      = "TERRAFORM_FOSS"
				}
			`, randomID),
				Check: Resource(
					"spacelift_module.workflow_tool_terraform_foss",
					Attribute("workflow_tool", Equals("TERRAFORM_FOSS")),
				),
			},
			// Can create a module with CUSTOM
			{
				Config: fmt.Sprintf(`
				resource "spacelift_module" "workflow_tool_custom" {
			        name               = "workflow-tool-custom-%s"
					branch             = "master"
					repository         = "terraform-bacon-tasty"
			        terraform_provider = "papaya"
					workflow_tool      = "CUSTOM"
				}
			`, randomID),
				Check: Resource(
					"spacelift_module.workflow_tool_custom",
					Attribute("workflow_tool", Equals("CUSTOM")),
				),
			},
		})
	})
}

func TestModuleResourceSpace(t *testing.T) {
	t.Run("with GitHub", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string, protectFromDeletion bool) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
					name                  = "github-module-%s"
					administrative        = true
					branch                = "master"
					description           = "%s"
					labels                = ["one", "two"]
					protect_from_deletion = %t
					repository            = "terraform-bacon-tasty"
					space_id              = "root"
					shared_accounts       = ["foo-subdomain", "bar-subdomain"]
				}
			`, randomID, description, protectFromDeletion)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description", true),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("terraform-default-github-module-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals(fmt.Sprintf("github-module-%s", randomID))),
					AttributeNotPresent("project_root"),
					Attribute("protect_from_deletion", Equals("true")),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("default")),
					Attribute("space_id", Equals("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description", false),
				Check: Resource(
					"spacelift_module.test",
					Attribute("description", Equals("new description")),
					Attribute("protect_from_deletion", Equals("false")),
				),
			},
		})
	})

	t.Run("project root and custom name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(projectRoot string) string {
			return fmt.Sprintf(`
				resource "spacelift_module" "test" {
                    name               = "project-root-%s"
					administrative     = true
					branch             = "master"
					description        = "description"
					labels             = ["one", "two"]
                    project_root       = "%s"
					repository         = "terraform-bacon-tasty"
					shared_accounts    = ["foo-subdomain", "bar-subdomain"]
                    terraform_provider = "papaya"
				}
			`, randomID, projectRoot)
		}

		const resourceName = "spacelift_module.test"

		testSteps(t, []resource.TestStep{
			{
				Config: config("test-root/ab"),
				Check: Resource(
					"spacelift_module.test",
					Attribute("id", Equals(fmt.Sprintf("terraform-papaya-project-root-%s", randomID))),
					Attribute("administrative", Equals("true")),
					Attribute("branch", Equals("master")),
					Attribute("description", Equals("description")),
					SetEquals("labels", "one", "two"),
					Attribute("name", Equals(fmt.Sprintf("project-root-%s", randomID))),
					Attribute("project_root", Equals("test-root/ab")),
					Attribute("repository", Equals("terraform-bacon-tasty")),
					SetEquals("shared_accounts", "bar-subdomain", "foo-subdomain"),
					Attribute("terraform_provider", Equals("papaya")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("test-root/bc"),
				Check:  Resource("spacelift_module.test", Attribute("project_root", Equals("test-root/bc"))),
			},
		})
	})

	for _, name := range []string{
		"github-Module",
		"github-module-",
		"_github-module",
		"0github-module",
	} {
		t.Run("invalid name", func(t *testing.T) {
			testSteps(t, []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						resource "spacelift_module" "test" {
							name                  = "%s"
							branch                = "master"
							repository            = "terraform-bacon-tasty"
						}
			`, name),
					ExpectError: regexp.MustCompile("must start and end with lowercase letter and may only contain lowercase letters, digits, dashes and underscores"),
				},
			})
		})
	}

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_module" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = ["one", "two"]
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_module.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_module" "test" {
					name                  = "labelled-module-%s"
					branch                = "master"
					labels                = []
					repository            = "terraform-bacon-tasty"
				}`, randomID),
				Check: Resource(
					"spacelift_module.test",
					SetEquals("labels"),
				),
			},
		})
	})
}
