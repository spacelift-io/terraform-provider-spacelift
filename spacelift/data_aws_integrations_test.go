package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestAWSIntegrationsData(t *testing.T) {
	first := &structs.AWSIntegration{
		DurationSeconds:             1234,
		GenerateCredentialsInWorker: false,
		Labels:                      []string{"one", "two"},
		Name:                        acctest.RandStringFromCharSet(5, acctest.CharSetAlpha),
		RoleARN:                     "arn:aws:iam::039653571618:role/empty-test-role",
		Space:                       "root",
	}
	second := &structs.AWSIntegration{
		DurationSeconds:             4321,
		GenerateCredentialsInWorker: true,
		Labels:                      []string{"three", "four"},
		Name:                        acctest.RandStringFromCharSet(5, acctest.CharSetAlpha),
		RoleARN:                     "arn:aws:iam::039653571618:role/empty-test-role-2",
		Space:                       "legacy",
	}

	terraformConfig := fmt.Sprintf(`
		 %s

		 %s

      	 data "spacelift_aws_integrations" "test" {
			depends_on = [spacelift_aws_integration.%s, spacelift_aws_integration.%s]
      	 }
		`,
		awsIntegrationToResource(first),
		awsIntegrationToResource(second),
		first.Name, second.Name,
	)
	testSteps(t, []resource.TestStep{{
		Config: terraformConfig, Check: resource.ComposeTestCheckFunc(
			Resource("data.spacelift_aws_integrations.test", Attribute("id", Equals("spacelift_aws_integrations"))),
			resource.ComposeTestCheckFunc(awsIntegrationChecks(first)...),
			resource.ComposeTestCheckFunc(awsIntegrationChecks(second)...),
		),
	}})
}

func awsIntegrationToResource(i *structs.AWSIntegration) string {
	return fmt.Sprintf(`
 resource "spacelift_aws_integration" "%s" {
        	name                           = "%s"
        	role_arn                       = "%s"	
			space_id 					   = "%s" 
        	labels                         =  %s
        	duration_seconds               =  %d
        	generate_credentials_in_worker =  %t
      	 }
`, i.Name, i.Name, i.RoleARN, i.Space, labelsAsString(i.Labels), i.DurationSeconds, i.GenerateCredentialsInWorker)
}

func labelsAsString(labels []string) string {
	return fmt.Sprintf(`["%s"]`, strings.Join(labels, `", "`))
}

func awsIntegrationChecks(first *structs.AWSIntegration) []resource.TestCheckFunc {
	resourceName := fmt.Sprintf("spacelift_aws_integration.%s", first.Name)
	return []resource.TestCheckFunc{
		Resource(resourceName, Attribute("name", Equals(first.Name))),
		Resource(resourceName, Attribute("role_arn", Equals(first.RoleARN))),
		Resource(resourceName, Attribute("space_id", Equals(first.Space))),
		Resource(resourceName, Attribute("duration_seconds", Equals(fmt.Sprintf("%d", first.DurationSeconds)))),
		Resource(resourceName, Attribute("generate_credentials_in_worker", Equals(fmt.Sprintf("%t", first.GenerateCredentialsInWorker)))),
		Resource(resourceName, SetEquals("labels", first.Labels...)),
	}
}
