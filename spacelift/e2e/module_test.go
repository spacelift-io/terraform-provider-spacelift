package e2e

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type ModuleTest struct {
	ResourceTest
}

func (e *ModuleTest) TestLifecycle_OK() {
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
		`{"query":"mutation($input:ModuleCreateInput!){moduleCreate(input: $input){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"input":{"updateInput":{"administrative":true,"branch":"master","description":"My description","labels":["label"],"workerPool":null},"namespace":null,"provider":"GITHUB","repository":"core-infra"}}}`,
		`{"data":{"moduleCreate":{"id":"babys-first-module"}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){module(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"module":{"id":"babys-first-module","administrative":true,"branch":"master","description":"My description","integrations":{"aws":{"assumeRolePolicyStatement":"bacon"},"gcp":{"serviceAccountEmail":null,"tokenScopes":[]},"webhooks":[]},"labels":["label"],"namespace":"terraform-super-module","provider":"GITHUB","repository":"core-infra","workerPool":null}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackDelete(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"stackDelete":{}}}`,
		1,
	)

	e.peekBody()

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_module" "module" {
	administrative    = true
	branch            = "master"
	description       = "My description"
	labels            = ["label"]
	namespace         = "terraform-super-module"
	repository        = "core-infra"
}

data "spacelift_module" "module" {
  module_id = spacelift_module.module.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_module.module", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_module.module", "administrative", "true"),
				resource.TestCheckResourceAttr("spacelift_module.module", "branch", "master"),
				resource.TestCheckResourceAttr("spacelift_module.module", "description", "My description"),
				resource.TestCheckResourceAttr("spacelift_module.module", "gitlab.#", "0"),
				resource.TestCheckResourceAttr("spacelift_module.module", "labels.#", "1"),
				resource.TestCheckResourceAttr("spacelift_module.module", "labels.3453406131", "label"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_module.module", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "administrative", "true"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "branch", "master"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "description", "My description"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "gitlab.#", "0"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "labels.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "repository", "core-infra"),
			),
		},
	})
}

func (e *ModuleTest) TestLifecycleGitlab_OK() {
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
		`{"query":"mutation($input:ModuleCreateInput!){moduleCreate(input: $input){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"input":{"updateInput":{"administrative":true,"branch":"master","description":"My description","labels":["label"],"workerPool":"worker-pool"},"namespace":"spacelift","provider":"GITLAB","repository":"core-infra"}}}`,
		`{"data":{"moduleCreate":{"id":"babys-first-module"}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){module(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"module":{"id":"babys-first-module","administrative":true,"branch":"master","description":"My description","integrations":{"aws":{"assumeRolePolicyStatement":"bacon"},"gcp":{"serviceAccountEmail":null,"tokenScopes":[]},"webhooks":[]},"labels":["label"],"namespace":"spacelift","provider":"GITLAB","repository":"core-infra","workerPool":{"id":"worker-pool"}}}}`,

		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackDelete(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"stackDelete":{}}}`,
		1,
	)

	e.peekBody()

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_module" "module" {
	gitlab {
		namespace = "spacelift"
	}

	administrative    = true
	branch            = "master"
	description       = "My description"
	labels            = ["label"]
	repository        = "core-infra"
	worker_pool_id    = "worker-pool"
}

data "spacelift_module" "module" {
  module_id = spacelift_module.module.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_module.module", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_module.module", "administrative", "true"),
				// resource.TestCheckResourceAttr("spacelift_module.module", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("spacelift_module.module", "branch", "master"),
				resource.TestCheckResourceAttr("spacelift_module.module", "description", "My description"),
				resource.TestCheckResourceAttr("spacelift_module.module", "gitlab.#", "1"),
				resource.TestCheckResourceAttr("spacelift_module.module", "gitlab.0.namespace", "spacelift"),
				resource.TestCheckResourceAttr("spacelift_module.module", "labels.#", "1"),
				resource.TestCheckResourceAttr("spacelift_module.module", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("spacelift_module.module", "repository", "core-infra"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_module.module", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "administrative", "true"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "aws_assume_role_policy_statement", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "branch", "master"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "description", "My description"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "gitlab.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "gitlab.0.namespace", "spacelift"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "labels.#", "1"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "labels.3453406131", "label"),
				resource.TestCheckResourceAttr("data.spacelift_module.module", "repository", "core-infra"),
			),
		},
	})
}

func TestModule(t *testing.T) {
	suite.Run(t, new(ModuleTest))
}
