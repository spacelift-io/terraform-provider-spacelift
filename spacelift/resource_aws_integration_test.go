package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_aws_integration.test"

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	testSteps(t, []resource.TestStep{
		{
			Config: fmt.Sprintf(`
			resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        generate_credentials_in_worker = false
			}
		`, randomID),
			Check: Resource(
				resourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("duration_seconds", Equals("900")),
				Attribute("generate_credentials_in_worker", Equals("false")),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				SetEquals("labels", "one", "two"),
			),
		},
		{
			ResourceName:      resourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: fmt.Sprintf(`
			resource "spacelift_aws_integration" "test" {
        name                           = "test-aws-integration-%s"
        role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        labels                         = ["one", "two"]
        duration_seconds               = 6000
        external_id                    = "external_id"
        generate_credentials_in_worker = true
			}
			`, randomID),
			Check: Resource(
				resourceName,
				Attribute("id", IsNotEmpty()),
				Attribute("duration_seconds", Equals("6000")),
				Attribute("external_id", Equals("external_id")),
				Attribute("generate_credentials_in_worker", Equals("true")),
				Attribute("role_arn", Equals("arn:aws:iam::039653571618:role/empty-test-role")),
				Attribute("name", Equals(fmt.Sprintf("test-aws-integration-%s", randomID))),
				SetEquals("labels", "one", "two"),
			),
		},
	})
}
