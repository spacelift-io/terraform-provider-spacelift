package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type ContextAttachmentTest struct {
	ResourceTest
}

func (e *ContextAttachmentTest) TestLifecycle_StackOK() {
	defer gock.Off()

	e.posts( // Mocking out the context attachment mutation.
		`{"query":"mutation($id:ID!$priority:Int!$stack:ID!){contextAttach(id: $id, stack: $stack, priority: $priority){id,stackId,isModule,priority}}","variables":{"id":"babys-first-context","priority":8,"stack":"babys-first-stack"}}`,
		`{"data":{"contextAttach":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}}`,
		1,
	)

	e.posts( // Mocking out the context attachment query.
		`{"query":"query($context:ID!$id:ID!){context(id: $context){attachedStack(id: $id){id,stackId,isModule,priority}}}","variables":{"context":"babys-first-context","id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"context":{"attachedStack":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","stackId":"babys-first-stack","priority":8}}}}`,
		5,
	)

	e.posts( // Mocking out the context detachment mutation.
		`{"query":"mutation($id:ID!){contextDetach(id: $id){id,stackId,isModule,priority}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"contextDetach":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_context_attachment" "attachment" {
  context_id = "babys-first-context"
  stack_id   = "babys-first-stack"
  priority   = 8
}

data "spacelift_context_attachment" "attachment" {
  attachment_id = spacelift_context_attachment.attachment.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "id", "babys-first-context/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "priority", "8"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "id", "babys-first-context/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "priority", "8"),
			),
		},
	})
}

func (e *ContextAttachmentTest) TestLifecycle_ModuleOK() {
	defer gock.Off()

	gock.Observe(gock.DumpRequest)

	e.posts( // Mocking out the context attachment mutation.
		`{"query":"mutation($id:ID!$priority:Int!$stack:ID!){contextAttach(id: $id, stack: $stack, priority: $priority){id,stackId,isModule,priority}}","variables":{"id":"babys-first-context","priority":8,"stack":"babys-first-module"}}`,
		`{"data":{"contextAttach":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}}`,
		1,
	)

	e.posts( // Mocking out the context attachment query.
		`{"query":"query($context:ID!$id:ID!){context(id: $context){attachedStack(id: $id){id,stackId,isModule,priority}}}","variables":{"context":"babys-first-context","id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"context":{"attachedStack":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV","stackId":"babys-first-module","isModule":true,"priority":8}}}}`,
		5,
	)

	e.posts( // Mocking out the context detachment mutation.
		`{"query":"mutation($id:ID!){contextDetach(id: $id){id,stackId,isModule,priority}}","variables":{"id":"01DJN6A8MHD9ZKYJ3NHC5QAPTV"}}`,
		`{"data":{"contextDetach":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_context_attachment" "attachment" {
  context_id = "babys-first-context"
  module_id  = "babys-first-module"
  priority   = 8
}

data "spacelift_context_attachment" "attachment" {
  attachment_id = spacelift_context_attachment.attachment.id
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "id", "babys-first-context/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_context_attachment.attachment", "priority", "8"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "id", "babys-first-context/01DJN6A8MHD9ZKYJ3NHC5QAPTV"),
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_context_attachment.attachment", "priority", "8"),
			),
		},
	})
}

func TestContextAttachment(t *testing.T) {
	suite.Run(t, new(ContextAttachmentTest))
}
