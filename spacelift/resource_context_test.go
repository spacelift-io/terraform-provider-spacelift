package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextResource(t *testing.T) {
	const resourceName = "spacelift_context.test"

	t.Run("creates and updates contexts without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					description = "%s"
					labels = ["one", "two"]
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-context-")),
					Attribute("name", StartsWith("Provider test context")),
					Attribute("description", Equals("old description")),
					SetEquals("labels", "one", "two"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					labels = ["one", "two"]
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					labels = []
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}

func TestContextResourceSpace(t *testing.T) {
	const resourceName = "spacelift_context.test"

	t.Run("creates and updates contexts without an error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					description = "%s"
					space_id  = "root"
					labels = ["one", "two"]
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("provider-test-context-")),
					Attribute("name", StartsWith("Provider test context")),
					Attribute("description", Equals("old description")),
					Attribute("space_id", Equals("root")),
					SetEquals("labels", "one", "two"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("new description"),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					labels = ["one", "two"]
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_context" "test" {
					name        = "Provider test context %s"
					labels = []
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}
