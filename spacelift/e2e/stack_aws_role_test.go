package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type StackAWSRoleTest struct {
	ResourceTest
}

func (e *StackAWSRoleTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$role_arn:String){stackSetAwsRoleDelegation(id: $id, roleArn: $roleArn){id,administrative,awsAssumedRoleARN,branch,description,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-stack","role_arn":"arn:aws:iam::075108987694:role/terraform"}}`,
		`{"data":{"stackSetAwsRoleDelegation":{}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,awsAssumedRoleARN,branch,description,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"awsAssumedRoleARN":"arn:aws:iam::075108987694:role/terraform"}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!$role_arn:String){stackSetAwsRoleDelegation(id: $id, roleArn: $roleArn){id,administrative,awsAssumedRoleARN,branch,description,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-stack","role_arn":null}}`,
		`{"data":{"stackSetAwsRoleDelegation":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_stack_aws_role" "role" {
  stack_id = "babys-first-stack"
  role_arn = "arn:aws:iam::075108987694:role/terraform"
}

data "spacelift_stack_aws_role" "role" {
  stack_id = spacelift_stack_aws_role.role.stack_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack_aws_role.role", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("spacelift_stack_aws_role.role", "stack_id", "babys-first-stack"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack_aws_role.role", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("data.spacelift_stack_aws_role.role", "stack_id", "babys-first-stack"),
			),
		},
	})
}

func TestStackAWSRole(t *testing.T) {
	suite.Run(t, new(StackAWSRoleTest))
}
