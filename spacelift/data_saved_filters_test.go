package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

// Note: Tests could interfere with each other if run in parallel.
// It's not a problem if we use different types (stacks, contexts, ...).
// But if we use the same type, tests could fail (for example listing saved filters - ).

func TestSavedFiltersData(t *testing.T) {
	t.Run("load all saved filters", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		datasourceName := "data.spacelift_saved_filters.all"

		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_saved_filter" "test" {
					name = "test-` + randomID + `"		
					type = "contexts"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "all" {
					depends_on = [spacelift_saved_filter.test]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName, Attribute("id", IsNotEmpty())),
			),
		}})
	})

	t.Run("type specified", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		datasourceName := "data.spacelift_saved_filters.stacks"

		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_saved_filter" "test1" {
					name = "a-` + randomID + `"
					type = "stacks"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test2" {
					name = "b-` + randomID + `"
					type = "stacks"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test3" {
					name = "c-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "stacks" {
					filter_type = "stacks"
					depends_on = [spacelift_saved_filter.test1,spacelift_saved_filter.test2,spacelift_saved_filter.test3]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("filters.#", Equals("2")),
					Attribute("filters.0.id", IsNotEmpty()),
					Attribute("filters.0.type", Equals("stacks")),
					Attribute("filters.0.name", Equals("a-"+randomID)),
					Attribute("filters.1.id", IsNotEmpty()),
					Attribute("filters.1.type", Equals("stacks")),
					Attribute("filters.1.name", Equals("b-"+randomID)),
				),
			),
		}})
	})

	t.Run("name specified", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		datasourceName := "data.spacelift_saved_filters.blueprints"

		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_saved_filter" "test1" {
					name = "d-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test2" {
					name = "e-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test3" {
					name = "f-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "blueprints" {
					filter_name = "d-` + randomID + `"
					depends_on = [spacelift_saved_filter.test1,spacelift_saved_filter.test2,spacelift_saved_filter.test3]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("filters.#", Equals("1")),
					Attribute("filters.0.id", IsNotEmpty()),
					Attribute("filters.0.type", Equals("blueprints")),
					Attribute("filters.0.name", Equals("d-"+randomID)),
				),
			),
		}})
	})

	t.Run("type & name specified", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		datasourceName := "data.spacelift_saved_filters.blueprints"

		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_saved_filter" "test1" {
					name = "g-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test2" {
					name = "h-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test3" {
					name = "i-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "blueprints" {
					filter_type = "blueprints"
					filter_name = "h-` + randomID + `"
					depends_on = [spacelift_saved_filter.test1,spacelift_saved_filter.test2,spacelift_saved_filter.test3]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("filters.#", Equals("1")),
					Attribute("filters.0.id", IsNotEmpty()),
					Attribute("filters.0.type", Equals("blueprints")),
					Attribute("filters.0.name", Equals("h-"+randomID)),
				),
			),
		}})
	})

	t.Run("no results", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		datasourceName := "data.spacelift_saved_filters.blueprints"

		testSteps(t, []resource.TestStep{{
			Config: `
				resource "spacelift_saved_filter" "test1" {
					name = "j-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test2" {
					name = "k-` + randomID + `"
					type = "blueprints"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				resource "spacelift_saved_filter" "test3" {
					name = "l-` + randomID + `"
					type = "contexts"
					is_public = true
					data = jsonencode({
						"key": "activeFilters",
						"value": jsonencode({})
  					})
				}

				data "spacelift_saved_filters" "blueprints" {
					filter_type = "blueprints"
					filter_name = "l-` + randomID + `"
					depends_on = [spacelift_saved_filter.test1,spacelift_saved_filter.test2,spacelift_saved_filter.test3]
				}
			`,
			Check: resource.ComposeTestCheckFunc(
				Resource(datasourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("filters.#", Equals("0")),
				),
			),
		}})
	})
}
