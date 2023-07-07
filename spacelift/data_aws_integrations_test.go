package spacelift

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

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

func awsIntegrationChecks(i *structs.AWSIntegration) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		Resource("data.spacelift_aws_integrations.test",
			Nested("integrations",
				CheckInList(
					Attribute("name", Equals(i.Name)),
					Attribute("integration_id", IsNotEmpty()),
					Attribute("role_arn", Equals(i.RoleARN)),
					Attribute("space_id", Equals(i.Space)),
					Attribute("duration_seconds", Equals(fmt.Sprintf("%d", i.DurationSeconds))),
					Attribute("generate_credentials_in_worker", Equals(fmt.Sprintf("%t", i.GenerateCredentialsInWorker))),
					SetEquals("labels", i.Labels...),
				),
			),
		),
	}
}
