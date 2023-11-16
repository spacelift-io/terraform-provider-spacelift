package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

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
					after_apply = ["after_apply"]
					after_destroy = ["after_destroy"]
					after_init = ["after_init"]
					after_perform = ["after_perform"]
					after_plan = ["after_plan"]
					after_run = ["after_run"]
					before_apply = ["before_apply"]
					before_destroy = ["before_destroy"]
					before_init = ["before_init"]
					before_perform = ["before_perform"]
					before_plan = ["before_plan"]
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
					Attribute("after_apply.#", Equals("1")),
					Attribute("after_apply.0", Equals("after_apply")),
					Attribute("after_destroy.#", Equals("1")),
					Attribute("after_destroy.0", Equals("after_destroy")),
					Attribute("after_init.#", Equals("1")),
					Attribute("after_init.0", Equals("after_init")),
					Attribute("after_perform.#", Equals("1")),
					Attribute("after_perform.0", Equals("after_perform")),
					Attribute("after_plan.#", Equals("1")),
					Attribute("after_plan.0", Equals("after_plan")),
					Attribute("after_run.#", Equals("1")),
					Attribute("after_run.0", Equals("after_run")),
					Attribute("before_apply.#", Equals("1")),
					Attribute("before_apply.0", Equals("before_apply")),
					Attribute("before_destroy.#", Equals("1")),
					Attribute("before_destroy.0", Equals("before_destroy")),
					Attribute("before_init.#", Equals("1")),
					Attribute("before_init.0", Equals("before_init")),
					Attribute("before_perform.#", Equals("1")),
					Attribute("before_perform.0", Equals("before_perform")),
					Attribute("before_plan.#", Equals("1")),
					Attribute("before_plan.0", Equals("before_plan")),
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
