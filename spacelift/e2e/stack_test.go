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
		`{"query":"mutation($input:StackInput!$manageState:Boolean!$stackObjectID:String){stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"input":{"administrative":true,"autodeploy":true,"branch":"master","description":"My description","labels":["label"],"name":"Baby's first stack","namespace":null,"projectRoot":"/project","provider":"GITHUB","repository":"core-infra","terraformVersion":"0.12.6","workerPool":null},"manageState":true,"stackObjectID":"objectID"}}`,
		`{"data":{"stackCreate":{"id":"babys-first-stack"}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"id":"babys-first-stack","administrative":true,"autodeploy":true,"branch":"master","description":"My description","integrations":{"aws":{"assumeRolePolicyStatement":"bacon"},"gcp":{"serviceAccountEmail":null,"tokenScopes":[]},"webhooks":[]},"labels":["label"],"managesStateFile":true,"name":"Baby's first stack","namespace":"","projectRoot":"/project","provider":"GITHUB","repository":"core-infra","terraformVersion":"0.12.6","workerPool":null}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackDelete(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stackDelete":{}}}`,
		1,
	)

	e.peekBody()

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_stack" "stack" {
	administrative    = true
	autodeploy        = true
	branch            = "master"
	description       = "My description"
	import_state      = "bacon"
	labels            = ["label"]
	name              = "Baby's first stack"
	project_root      = "/project"
	repository        = "core-infra"
	terraform_version = "0.12.6"
}

data "spacelift_stack" "stack" {
  stack_id = spacelift_stack.stack.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "autodeploy", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "gitlab.#", "0"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "labels.#", "1"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "project_root", "/project"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "repository", "core-infra"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "terraform_version", "0.12.6"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "autodeploy", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "gitlab.#", "0"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "labels.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "repository", "core-infra"),
			),
		},
	})
}

func (e *StackTest) TestLifecycleGitlab_OK() {
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
		`{"query":"mutation($input:StackInput!$manageState:Boolean!$stackObjectID:String){stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"input":{"administrative":true,"autodeploy":true,"branch":"master","description":"My description","labels":["label"],"name":"Baby's first stack","namespace":"spacelift","projectRoot":"/project","provider":"GITLAB","repository":"core-infra","terraformVersion":"0.12.6","workerPool":"worker-pool"},"manageState":true,"stackObjectID":"objectID"}}`,
		`{"data":{"stackCreate":{"id":"babys-first-stack"}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"id":"babys-first-stack","administrative":true,"autodeploy":true,"branch":"master","description":"My description","integrations":{"aws":{"assumeRolePolicyStatement":"bacon"},"gcp":{"serviceAccountEmail":null,"tokenScopes":[]},"webhooks":[]},"labels":["label"],"managesStateFile":true,"name":"Baby's first stack","namespace":"spacelift","projectRoot":"/project","provider":"GITLAB","repository":"core-infra","terraformVersion":"0.12.6","workerPool":{"id":"worker-pool"}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackDelete(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stackDelete":{}}}`,
		1,
	)

	e.peekBody()

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_stack" "stack" {
	gitlab {
		namespace = "spacelift"
	}

	administrative    = true
	autodeploy        = true
	branch            = "master"
	description       = "My description"
	import_state      = "bacon"
	labels            = ["label"]
	name              = "Baby's first stack"
	project_root      = "/project"
	repository        = "core-infra"
	terraform_version = "0.12.6"
	worker_pool_id    = "worker-pool"
}

data "spacelift_stack" "stack" {
  stack_id = spacelift_stack.stack.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "autodeploy", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "gitlab.#", "1"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "gitlab.0.namespace", "spacelift"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "labels.#", "1"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "project_root", "/project"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "repository", "core-infra"),
				resource.TestCheckResourceAttr("spacelift_stack.stack", "terraform_version", "0.12.6"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "administrative", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "autodeploy", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "branch", "master"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "description", "My description"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "gitlab.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "gitlab.0.namespace", "spacelift"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "labels.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "manage_state", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack.stack", "repository", "core-infra"),
			),
		},
	})
}

func TestStack(t *testing.T) {
	suite.Run(t, new(StackTest))
}
