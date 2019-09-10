package e2e

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type StackTest struct {
	ResourceTest
}

func (e *StackTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation{stateUploadUrl{objectId,url}}"}`,
		`{"data":{"stateUploadUrl":{"objectId":"objectID","url":"http://bacon.org/upload"}}}`,
		1,
	)

	gock.
		New("http://bacon.org").
		Post("/upload").
		Times(1).
		BodyString("bacon").
		Reply(http.StatusOK)

	e.posts(
		`{"query":"mutation($input:StackInput!$manageState:bool$stackObjectID:String){stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID){id,administrative,awsAssumedRoleARN,awsAssumeRolePolicyStatement,branch,description,managesStateFile,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"input":{"administrative":true,"branch":"master","description":"My description","name":"Baby's first stack","readersSlug":"engineering","repo":"core-infra","terraformVersion":"0.12.6","writersSlug":"devops"},"manageState":true,"stackObjectID":"objectID"}}`,
		`{"data":{"stackCreate":{"id":"babys-first-stack"}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,awsAssumedRoleARN,awsAssumeRolePolicyStatement,branch,description,managesStateFile,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"id":"babys-first-stack","administrative":true,"awsAssumeRolePolicyStatement":"bacon","branch":"master","description":"My description","managesStateFile":true,"name":"Baby's first stack","readersSlug":"engineering","repo":"core-infra","terraformVersion":"0.12.6","writersSlug":"devops"}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:String!){stackDelete(id: $id){id,administrative,awsAssumedRoleARN,awsAssumeRolePolicyStatement,branch,description,managesStateFile,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stackDelete":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_stack" "stack" {
	administrative    = true
	branch            = "master"
	description       = "My description"
	import_state      = "bacon"
	name              = "Baby's first stack"
	readers_team      = "engineering"
	repository        = "core-infra"
	terraform_version = "0.12.6"
	writers_team      = "devops"
}

data "spacelift_stack" "stack" {
  stack_id = spacelift_stack.stack.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "readers_team", "engineering"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "repository", "core-infra"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "terraform_version", "0.12.6"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "writers_team", "devops"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "readers_team", "engineering"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "repository", "core-infra"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "terraform_version", "0.12.6"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "writers_team", "devops"),
			),
		},
	})
}

func TestStack(t *testing.T) {
	suite.Run(t, new(StackTest))
}
