package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestPolicyResource(t *testing.T) {
	const resourceName = "spacelift_policy.test"

	t.Run("creates and updates a policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					labels = ["one", "two"]
					description = "My awesome policy"
					body = <<EOF
					package spacelift
					deny["%s"] { true }
					EOF
					type = "PLAN"
					description = "My awesome policy"
				}
			`, randomID, message)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("boom"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("my-first-policy")),
					Attribute("body", Contains("boom")),
					Attribute("type", Equals("PLAN")),
					SetEquals("labels", "one", "two"),
					Attribute("description", Equals("My awesome policy")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("bang"),
				Check: Resource(
					resourceName,
					Attribute("body", Contains("bang")),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_policy" "test" {
					name = "Label test policy %s"
					labels = ["one", "two", "three"]
					body = "package spacelift"
					type = "PLAN"
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two", "three"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_policy" "test" {
					name = "Label test policy %s"
					labels = []
					body = "package spacelift"
					type = "PLAN"
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}

func TestPolicyResourceSpace(t *testing.T) {
	const resourceName = "spacelift_policy.test"

	t.Run("creates and updates a policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := func(message string) string {
			return fmt.Sprintf(`
				resource "spacelift_policy" "test" {
					name = "My first policy %s"
					labels = ["one", "two"]
					description = "My awesome policy"
					space_id = "root"
					body = <<EOF
					package spacelift

					deny["%s"] { true }
					EOF
					type = "PLAN"
					description = "My awesome policy"
				}
			`, randomID, message)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("boom"),
				Check: Resource(
					resourceName,
					Attribute("id", StartsWith("my-first-policy")),
					Attribute("body", Contains("boom")),
					Attribute("type", Equals("PLAN")),
					SetEquals("labels", "one", "two"),
					Attribute("space_id", Equals("root")),
					Attribute("description", Equals("My awesome policy")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("bang"),
				Check: Resource(
					resourceName,
					Attribute("body", Contains("bang")),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_policy" "test" {
					name = "Label test policy %s"
					labels = ["one", "two", "three"]
					body = "package spacelift"
					type = "PLAN"
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels", "one", "two", "three"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_policy" "test" {
					name = "Label test policy %s"
					labels = []
					body = "package spacelift"
					type = "PLAN"
				}`, randomID),
				Check: Resource(
					resourceName,
					SetEquals("labels"),
				),
			},
		})
	})
}
