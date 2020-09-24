package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type StackGCPServiceAccountTest struct {
	ResourceTest
}

func (e *StackGCPServiceAccountTest) TestLifecycle_Module() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$tokenScopes:[String!]!){stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes){activated}}","variables":{"id":"babys-first-module","tokenScopes":["bacon","cabbage"]}}`,
		`{"data":{"stackIntegrationGcpCreate":{"activated":true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){module(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"module":{"integrations":{"gcp":{"serviceAccountEmail":"email","tokenScopes":["bacon","cabbage"]}}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackIntegrationGcpDelete(id: $id){activated}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"stackIntegrationGcpDelete":{"activated":false}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_gcp_service_account" "account" {
  module_id = "babys-first-module"
  token_scopes = ["bacon","cabbage"]
}

data "spacelift_gcp_service_account" "account" {
  module_id = spacelift_gcp_service_account.account.module_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.2763263256", "bacon"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.3359016956", "cabbage"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.1891707634", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.3620376089", "cabbage"),
			),
		},
	})
}
func (e *StackGCPServiceAccountTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$tokenScopes:[String!]!){stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes){activated}}","variables":{"id":"babys-first-stack","tokenScopes":["bacon","cabbage"]}}`,
		`{"data":{"stackIntegrationGcpCreate":{"activated":true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"integrations":{"gcp":{"serviceAccountEmail":"email","tokenScopes":["bacon","cabbage"]}}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($id:ID!){stackIntegrationGcpDelete(id: $id){activated}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stackIntegrationGcpDelete":{"activated":false}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_gcp_service_account" "account" {
  stack_id = "babys-first-stack"
  token_scopes = ["bacon","cabbage"]
}

data "spacelift_gcp_service_account" "account" {
  stack_id = spacelift_gcp_service_account.account.stack_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.2763263256", "bacon"),
				resource.TestCheckResourceAttr("spacelift_gcp_service_account.account", "token_scopes.3359016956", "cabbage"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.1891707634", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_gcp_service_account.account", "token_scopes.3620376089", "cabbage"),
			),
		},
	})
}

func TestGCPServiceAccount(t *testing.T) {
	suite.Run(t, new(StackGCPServiceAccountTest))
}
