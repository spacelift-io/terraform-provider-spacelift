package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type PolicyTest struct {
	ResourceTest
}

func (e *PolicyTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts( // Mocking out the policy creation.
		`{"query":"mutation($body:String!$name:String!$type:PolicyType!){policyCreate(name: $name, body: $body, type: $type){id,name,body,type}}","variables":{"body":"package spacelift","name":"Baby's first policy","type":"STACK_ACCESS"}}`,
		`{"data":{"policyCreate":{"id":"babys-first-policy"}}}`,
		1,
	)

	e.posts( // Mocking out the policy query.
		`{"query":"query($id:ID!){policy(id: $id){id,name,body,type}}","variables":{"id":"babys-first-policy"}}`,
		`{"data":{"policy":{"id":"babys-first-policy","name":"Baby's first policy","body":"package spacelift","type":"STACK_ACCESS"}}}`,
		6,
	)

	e.posts( // Mocking out the policy deletion.
		`{"query":"mutation($id:String!){policyDelete(id: $id){id,name,body,type}}","variables":{"id":"babys-first-policy"}}`,
		`{"data":{"policyDelete":{"id":"babys-first-policy"}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{{
		Config: `
resource "spacelift_policy" "policy" {
  name = "Baby's first policy"
  body = "package spacelift"
  type = "STACK_ACCESS"
}

data "spacelift_policy" "policy" {
  policy_id = spacelift_policy.policy.id
}
`,
		Check: resource.ComposeTestCheckFunc(
			// Test resource.
			resource.TestCheckResourceAttr("spacelift_policy.policy", "id", "babys-first-policy"),
			resource.TestCheckResourceAttr("spacelift_policy.policy", "name", "Baby's first policy"),
			resource.TestCheckResourceAttr("spacelift_policy.policy", "body", "package spacelift"),
			resource.TestCheckResourceAttr("spacelift_policy.policy", "type", "STACK_ACCESS"),

			// Test data.
			resource.TestCheckResourceAttr("data.spacelift_policy.policy", "id", "babys-first-policy"),
			resource.TestCheckResourceAttr("data.spacelift_policy.policy", "name", "Baby's first policy"),
			resource.TestCheckResourceAttr("data.spacelift_policy.policy", "body", "package spacelift"),
			resource.TestCheckResourceAttr("data.spacelift_policy.policy", "type", "STACK_ACCESS"),
		),
	}})
}

func TestPolicy(t *testing.T) {
	suite.Run(t, new(PolicyTest))
}
