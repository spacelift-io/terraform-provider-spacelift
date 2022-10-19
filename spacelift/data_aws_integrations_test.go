package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationsData(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	resourceName := "spacelift_aws_integration.test"
	datasourceName := "data.spacelift_aws_integrations.test"

	testSteps(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
		 resource "spacelift_aws_integration" "test" {
        	name                           = "test-aws-integration-%s"
        	role_arn                       = "arn:aws:iam::039653571618:role/empty-test-role"
        	labels                         = ["one", "two"]
        	duration_seconds               = 3600
        	generate_credentials_in_worker = false
      	 }

      	 data "spacelift_aws_integrations" "test" {
			depends_on = [spacelift_aws_integration.test]
      	 }
		`, randomID), Check: resource.ComposeTestCheckFunc(
			Resource(datasourceName, Attribute("id", IsNotEmpty())),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"integrations", "role_arn"}, resourceName, "role_arn"),
			CheckIfResourceNestedAttributeContainsResourceAttribute(datasourceName, []string{"integrations", "name"}, resourceName, "name"),
		),
	}})
}
