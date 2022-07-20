package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

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
}
