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

func (e *StackGCPServiceAccountTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($id:ID!$tokenScopes:[String!]!){stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes){activated}}","variables":{"id":"babys-first-stack","tokenScopes":["bacon","cabbage"]}}`,
		`{"data":{"stackIntegrationGcpCreate":{"activated":true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,repository,terraformVersion}}","variables":{"id":"babys-first-stack"}}`,
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
resource "spacelift_stack_gcp_service_account" "account" {
  stack_id = "babys-first-stack"
  token_scopes = ["bacon","cabbage"]
}

data "spacelift_stack_gcp_service_account" "account" {
  stack_id = spacelift_stack_gcp_service_account.account.stack_id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "token_scopes.2763263256", "bacon"),
				resource.TestCheckResourceAttr("spacelift_stack_gcp_service_account.account", "token_scopes.3359016956", "cabbage"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "service_account_email", "email"),
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "token_scopes.#", "2"),
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "token_scopes.1891707634", "bacon"),
				resource.TestCheckResourceAttr("data.spacelift_stack_gcp_service_account.account", "token_scopes.3620376089", "cabbage"),
			),
		},
	})
}

func TestStackGCPServiceAccount(t *testing.T) {
	suite.Run(t, new(StackGCPServiceAccountTest))
}
