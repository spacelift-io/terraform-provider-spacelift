package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestContextsData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resourceName := "spacelift_context.test"
	datasourceName := "data.spacelift_contexts.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			locals {
				seed = "%s"
			}

			resource "spacelift_context" "test" {
				name   = "My first context ${local.seed}"
				labels = ["foo-${local.seed}", "bar-${local.seed}"]
			}

			resource "spacelift_context" "test2" {
				name   = "My second context ${local.seed}"
				labels = ["baz-${local.seed}", "qux-${local.seed}"]
			}

			data "spacelift_contexts" "test" {
				depends_on = [
					spacelift_context.test,
					spacelift_context.test2,
				]

				labels {
					any_of = ["foo-${local.seed}", "abc-${local.seed}"]
				}

				labels {
					any_of = ["bar-${local.seed}", "def-${local.seed}"]
				}
			}
		`, randomID), Check: resource.ComposeTestCheckFunc(
			Resource(datasourceName, Attribute("id", IsNotEmpty())),
			Resource(datasourceName, Attribute("contexts.#", Equals("1"))),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"contexts", "context_id"}, resourceName, "id"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"contexts", "name"}, resourceName, "name"),
		),
	}})
}

func TestContextsDataSpace(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resourceName := "spacelift_context.test"
	datasourceName := "data.spacelift_contexts.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
			resource "spacelift_context" "test" {
				name = "My first context %s"
				space_id = "root"
			}

			data "spacelift_contexts" "test" {
				depends_on = [spacelift_context.test]
			}
		`, randomID), Check: resource.ComposeTestCheckFunc(
			Resource(datasourceName, Attribute("id", IsNotEmpty())),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"contexts", "context_id"}, resourceName, "id"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"contexts", "name"}, resourceName, "name"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"contexts", "space_id"}, resourceName, "space_id"),
		),
	}})
}
