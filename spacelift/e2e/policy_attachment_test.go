package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type PolicyAttachmentTest struct {
	ResourceTest
}

func (e *PolicyAttachmentTest) TestLifecycle_Module() {
	defer gock.Off()

	e.posts( // Mocking out the policy attachment mutation.
		`{"query":"mutation($customInput:String$id:ID!$stack:ID!){policyAttach(id: $id, stack: $stack, customInput: $customInput){id,stackId,isModule,customInput}}","variables":{"id":"babys-first-policy","stack":"babys-first-module","customInput":"{\"bacon\":\"tasty\"}"}}`,
		`{"data":{"policyAttach":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}}`,
		1,
	)

	e.posts( // Mocking out the policy attachment query.
		`{"query":"query($id:ID!$policy:ID!){policy(id: $policy){attachedStack(id: $id){id,stackId,isModule,customInput}}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","policy":"babys-first-policy"}}`,
		`{"data":{"policy":{"attachedStack":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","stackId":"babys-first-module","isModule":true,"customInput":"{\"bacon\":\"tasty\"}"}}}}}}`,
		2,
	)

	e.posts( // Mocking out the policy detachment mutation.
		`{"query":"mutation($id:ID!){policyDetach(id: $id){id,stackId,isModule,customInput}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"policyDetach":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_policy_attachment" "attachment" {
  policy_id    = "babys-first-policy"
  module_id    = "babys-first-module"
  custom_input = jsonencode({ bacon = "tasty" })
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "id", "babys-first-policy/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "custom_input", `{"bacon":"tasty"}`),
				resource.TestCheckNoResourceAttr("spacelift_policy_attachment.attachment", "stack_id"),
			),
		},
	})
}

func (e *PolicyAttachmentTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts( // Mocking out the policy attachment mutation.
		`{"query":"mutation($customInput:String$id:ID!$stack:ID!){policyAttach(id: $id, stack: $stack, customInput: $customInput){id,stackId,isModule,customInput}}","variables":{"id":"babys-first-policy","stack":"babys-first-stack","customInput":"{\"bacon\":\"tasty\"}"}}`,
		`{"data":{"policyAttach":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}}`,
		1,
	)

	e.posts( // Mocking out the policy attachment query.
		`{"query":"query($id:ID!$policy:ID!){policy(id: $policy){attachedStack(id: $id){id,stackId,isModule,customInput}}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","policy":"babys-first-policy"}}`,
		`{"data":{"policy":{"attachedStack":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","stackId":"babys-first-stack","isModule":false,"customInput":"{\"bacon\":\"tasty\"}"}}}}`,
		2,
	)

	e.posts( // Mocking out the policy detachment mutation.
		`{"query":"mutation($id:ID!){policyDetach(id: $id){id,stackId,isModule,customInput}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"policyDetach":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_policy_attachment" "attachment" {
  policy_id    = "babys-first-policy"
  stack_id     = "babys-first-stack"
  custom_input = jsonencode({ bacon = "tasty" })
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "id", "babys-first-policy/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_policy_attachment.attachment", "custom_input", `{"bacon":"tasty"}`),
				resource.TestCheckNoResourceAttr("spacelift_policy_attachment.attachment", "module_id"),
			),
		},
	})
}

func TestPolicyAttachment(t *testing.T) {
	suite.Run(t, new(PolicyAttachmentTest))
}
