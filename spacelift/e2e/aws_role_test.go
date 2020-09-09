package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type AWSRoleTest struct {
	ResourceTest
}

func (e *AWSRoleTest) TestLifecycle_Module() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$roleArn:String!){stackIntegrationAwsCreate(id: $id, roleArn: $roleArn){activated}}","variables":{"id":"babys-first-module","roleArn":"arn:aws:iam::075108987694:role/terraform"}}`,
		`{"data":{"stackIntegrationAwsCreate":{"activated":true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){module(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"module":{"integrations":{"aws":{"assumedRoleArn":"arn:aws:iam::075108987694:role/terraform"}}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackIntegrationAwsDelete(id: $id){activated}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"stackIntegrationAwsDelete":{"activated":false}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_aws_role" "role" {
  module_id = "babys-first-module"
  role_arn = "arn:aws:iam::075108987694:role/terraform"
}

data "spacelift_aws_role" "role" {
  module_id = spacelift_aws_role.role.module_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "module_id", "babys-first-module"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "module_id", "babys-first-module"),
			),
		},
	})
}

func (e *AWSRoleTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$roleArn:String!){stackIntegrationAwsCreate(id: $id, roleArn: $roleArn){activated}}","variables":{"id":"babys-first-stack","roleArn":"arn:aws:iam::075108987694:role/terraform"}}`,
		`{"data":{"stackIntegrationAwsCreate":{"activated":true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"integrations":{"aws":{"assumedRoleArn":"arn:aws:iam::075108987694:role/terraform"}}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackIntegrationAwsDelete(id: $id){activated}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stackIntegrationAwsDelete":{"activated":false}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_aws_role" "role" {
  stack_id = "babys-first-stack"
  role_arn = "arn:aws:iam::075108987694:role/terraform"
}

data "spacelift_aws_role" "role" {
  stack_id = spacelift_aws_role.role.stack_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("spacelift_aws_role.role", "stack_id", "babys-first-stack"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "role_arn", "arn:aws:iam::075108987694:role/terraform"),
				resource.TestCheckResourceAttr("data.spacelift_aws_role.role", "stack_id", "babys-first-stack"),
			),
		},
	})
}

func TestStackAWSRole(t *testing.T) {
	suite.Run(t, new(AWSRoleTest))
}
