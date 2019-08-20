package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type MountedFileTest struct {
	ResourceTest
}

func (e *MountedFileTest) TestLifecycle_Context() {
	defer gock.Off()

	e.posts(
		`{"query":"query($context:ID!$id:ID!){context(id: $context){configElement(id: $id){id,checksum,type,value,writeOnly}}}","variables":{"context":"babys-first-context","id":"secret/key"}}`,
		`{"data":{"context":{"configElement":{"id":"context/babys-first-context/secret/key","checksum":"916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94","type":"FILE_MOUNT","value":null,"writeOnly":true}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($config:ConfigInput!$context:ID!){contextConfigAdd(context: $id, config: $input){id,checksum,type,value,writeOnly}}","variables":{"config":{"id":"secret/key","type":"FILE_MOUNT","value":"YmFjb24=","writeOnly":true},"context":"babys-first-context"}}`,
		`{"data":{"contextConfigAdd":{"id":"context/babys-first-context/secret/key"}}}`,
		1,
	)

	e.posts(
		`{"query":"mutation($context:ID!$id:ID!){contextConfigDelete(context: $context, id: $id){id,checksum,type,value,writeOnly}}","variables":{"context":"babys-first-context","id":"secret/key"}}`,
		`{"data":{"contextConfigDelete":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_mounted_file" "context_file" {
  context_id    = "babys-first-context"
  relative_path = "secret/key"
  content       = "YmFjb24="
}

data "spacelift_mounted_file" "context_file" {
  context_id    = "babys-first-context"
  relative_path = "secret/key"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "id", "context/babys-first-context/secret/key"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "checksum", "916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "content", ""),
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "context_id", "babys-first-context"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "relative_path", "secret/key"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.context_file", "write_only", "true"),
				resource.TestCheckNoResourceAttr("spacelift_mounted_file.context_file", "stack_id"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "id", "context/babys-first-context/secret/key"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "checksum", "916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "content", ""),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "context_id", "babys-first-context"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "relative_path", "secret/key"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.context_file", "write_only", "true"),
				resource.TestCheckNoResourceAttr("data.spacelift_mounted_file.context_file", "stack_id"),
			),
		},
	})
}

func (e *MountedFileTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts(
		`{"query":"query($id:ID!$stack:ID!){stack(id: $stack){configElement(id: $id){id,checksum,type,value,writeOnly}}}","variables":{"id":"secret/key","stack":"babys-first-stack"}}`,
		`{"data":{"stack":{"configElement":{"id":"stack/babys-first-stack/secret/key","checksum":"916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94","type":"FILE_MOUNT","value":null,"writeOnly":true}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($config:ConfigInput!$stack:ID!){stackConfigAdd(stack: $id, config: $input){id,checksum,type,value,writeOnly}}","variables":{"config":{"id":"secret/key","type":"FILE_MOUNT","value":"YmFjb24=","writeOnly":true},"stack":"babys-first-stack"}}`,
		`{"data":{"stackConfigAdd":{"id":"stack/babys-first-stack/secret/key"}}}`,
		1,
	)

	e.posts(
		`{"query":"mutation($id:ID!$stack:ID!){stackConfigDelete(stack: $stack, id: $id){id,checksum,type,value,writeOnly}}","variables":{"id":"secret/key","stack":"babys-first-stack"}}`,
		`{"data":{"stackConfigDelete":{}}}`,
		1,
	)

	e.peekBody()

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_mounted_file" "stack_file" {
  stack_id      = "babys-first-stack"
  relative_path = "secret/key"
  content       = "YmFjb24="
}

data "spacelift_mounted_file" "stack_file" {
  stack_id      = "babys-first-stack"
  relative_path = "secret/key"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "id", "stack/babys-first-stack/secret/key"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "checksum", "916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "relative_path", "secret/key"),
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "content", ""),
				resource.TestCheckResourceAttr("spacelift_mounted_file.stack_file", "write_only", "true"),
				resource.TestCheckNoResourceAttr("spacelift_mounted_file.stack_file", "context_id"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "id", "stack/babys-first-stack/secret/key"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "checksum", "916de3d556bb83674e3a297e77a6ba4a590f28b15c119640d4a22a31c6182d94"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "relative_path", "secret/key"),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "content", ""),
				resource.TestCheckResourceAttr("data.spacelift_mounted_file.stack_file", "write_only", "true"),
				resource.TestCheckNoResourceAttr("data.spacelift_mounted_file.stack_file", "context_id"),
			),
		},
	})
}

func TestMountedFile(t *testing.T) {
	suite.Run(t, new(MountedFileTest))
}
