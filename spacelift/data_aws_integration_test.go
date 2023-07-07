package spacelift

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationData(t *testing.T) {
	t.Run("without generating AWS creds in the worker", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
      resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 3600
        generate_credentials_in_worker = false
      }

      data "spacelift_aws_integration" "test" {
        integration_id = spacelift_aws_integration.test.id
      }
      `, randomID),
			Check: Resource(
				"data.spacelift_aws_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("duration_seconds", Equals("3600")),
				Attribute("generate_credentials_in_worker", Equals("false")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				SetEquals("labels", "one", "two"),
			),
		}})
	})

	t.Run("with generating AWS creds in the worker", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
      resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 6000
        external_id                    = "external_id"
        generate_credentials_in_worker = true
      }

      data "spacelift_aws_integration" "test" {
        integration_id = spacelift_aws_integration.test.id
      }
      `, randomID),
			Check: Resource(
				"data.spacelift_aws_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("duration_seconds", Equals("6000")),
				Attribute("generate_credentials_in_worker", Equals("true")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				Attribute("external_id", Equals("external_id")),
				SetEquals("labels", "one", "two"),
			),
		}})
	})

	t.Run("can lookup the integration by name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
      resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 3600
        generate_credentials_in_worker = false
      }

      data "spacelift_aws_integration" "test" {
        name = spacelift_aws_integration.test.name
      }
      `, randomID),
			Check: Resource(
				"data.spacelift_aws_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("duration_seconds", Equals("3600")),
				Attribute("generate_credentials_in_worker", Equals("false")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				SetEquals("labels", "one", "two"),
			),
		}})
	})

	t.Run("when integration ID does not exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_aws_integration" "test" {
				integration_id = "01GBASTWAEPJ1HDMXDMWTRC8DN"
			}`,
			ExpectError: regexp.MustCompile(`AWS integration not found: 01GBASTWAEPJ1HDMXDMWTRC8DN`),
		}})
	})

	t.Run("when integration name does not exist", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
			data "spacelift_aws_integration" "test" {
				name = "non-existent integration"
			}`,
			ExpectError: regexp.MustCompile(`AWS integration not found: non-existent integration`),
		}})
	})

	t.Run("when setting both integration_id and name it errors", func(t *testing.T) {
		testSteps(t, []resource.TestStep{{
			Config: `
	data "spacelift_aws_integration" "test" {
        integration_id = "01GBAME4P2BS72ZQRA9HJYWRCK"
		name           = "Test Integration"
    }
    `,
			ExpectError: regexp.MustCompile("only one of `integration_id,name` can be specified"),
		}})
	})
}

func TestAWSIntegrationDataSpace(t *testing.T) {
	t.Run("without generating AWS creds in the worker", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
      resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 3600
        generate_credentials_in_worker = false
        space_id                       = "root"
      }

      data "spacelift_aws_integration" "test" {
        integration_id = spacelift_aws_integration.test.id
      }
      `, randomID),
			Check: Resource(
				"data.spacelift_aws_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("duration_seconds", Equals("3600")),
				Attribute("space_id", Equals("root")),
				Attribute("generate_credentials_in_worker", Equals("false")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				SetEquals("labels", "one", "two"),
			),
		}})
	})

	t.Run("with generating AWS creds in the worker", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
      resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 6000
        external_id                    = "external_id"
        space_id                       = "root"
        generate_credentials_in_worker = true
      }

      data "spacelift_aws_integration" "test" {
        integration_id = spacelift_aws_integration.test.id
      }
      `, randomID),
			Check: Resource(
				"data.spacelift_aws_integration.test",
				Attribute("id", IsNotEmpty()),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("duration_seconds", Equals("6000")),
				Attribute("generate_credentials_in_worker", Equals("true")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				Attribute("external_id", Equals("external_id")),
				Attribute("space_id", Equals("root")),
				SetEquals("labels", "one", "two"),
			),
		}})
	})
}
