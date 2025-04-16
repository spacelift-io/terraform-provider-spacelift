package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBlueprintVersionResource(t *testing.T) {
	const resourceName = "spacelift_blueprint_version.test"
	const resourceNameSecond = "spacelift_blueprint_version.test2"

	t.Run("Creates multiple and updates a blueprint in DRAFT state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_blueprint_versioned_group" "test" {
					name        = "test-blueprint-group-%s"
					space       = "root"
					description = "this is a test group"
				}

				resource "spacelift_blueprint_version" "test" {
					group 		= spacelift_blueprint_versioned_group.test.id 
					description = "%s"
					labels      = ["one", "two"]
					state       = "DRAFT"
					template    = "not validated for drafts"
					version 	= "1.0.0"
				}
				resource "spacelift_blueprint_version" "test2" {
					group 		= spacelift_blueprint_versioned_group.test.id 
					description = "second version"
					state       = "DRAFT"
					template    = "not validated for drafts2"
					version 	= "1.0.1"
				}`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("test description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.0")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("test description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts")),
				),
			},
			{
				Config: config("second version"),
				Check: Resource(
					resourceNameSecond,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.1")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("second version")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts2")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("updated description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.0")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("updated description")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts")),
				),
			},
			{
				Config: config("second unchanged"),
				Check: Resource(
					resourceNameSecond,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.1")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("second version")),
					Attribute("state", Equals("DRAFT")),
					Attribute("template", Equals("not validated for drafts2")),
				),
			},
		})
	})

	t.Run("Creates and updates a blueprint in PUBLISHED state", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		validTemplate1 := `stack:\n  name: stackerino\n  space: root\n  vcs:\n    branch: main\n    repository: spacelift-io/terraform-provider-spacelift\n    provider: GITHUB\n  vendor:\n    terraform:\n      manage_state: true\n      version: 0.12.0`
		validTemplate2 := `stack:\n  name: stackerino\n  space: root\n  vcs:\n    branch: main\n    repository: spacelift-io/terraform-provider-spacelift\n    provider: GITHUB\n  vendor:\n    terraform:\n      manage_state: true\n      version: 0.13.0`

		config := func(template, description string) string {
			return fmt.Sprintf(`
				resource "spacelift_blueprint_versioned_group" "test" {
					name        = "test-blueprint-group-%s"
					space       = "root"
					description = "this is a test group"
				}

				resource "spacelift_blueprint_version" "test" {
					group 		= spacelift_blueprint_versioned_group.test.id 
					description = "%s"
					labels      = ["one", "two"]
					state       = "PUBLISHED"
					template    = "%s"
					version 	= "1.0.0"
				}`, randomID, description, template)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(validTemplate1, "test description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.0")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("test description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("PUBLISHED")),
					Attribute("template", IsNotEmpty()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(validTemplate2, "updated description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("version", Equals("1.0.0")),
					Attribute("group", Equals("test-blueprint-group-"+randomID)),
					Attribute("description", Equals("updated description")),
					Attribute("labels.#", Equals("2")),
					Attribute("state", Equals("PUBLISHED")),
					Attribute("template", IsNotEmpty()),
				),
			},
		})
	})
}
