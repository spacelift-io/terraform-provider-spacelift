package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type EnvironmentVariableTest struct {
	ResourceTest
}

func (e *EnvironmentVariableTest) TestLifecycle_Context() {
	defer gock.Off()

	e.posts(
		`{"query":"query($context:ID!$id:ID!){context(id: $context){configElement(id: $id){id,checksum,type,value,writeOnly}}}","variables":{"context":"babys-first-context","id":"SECRET_KEY"}}`,
		`{"data":{"context":{"configElement":{"id":"context/babys-first-context/SECRET_KEY","checksum":"f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1","type":"ENVIRONMENT_VARIABLE","value":null,"writeOnly":true}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($config:ConfigInput!$context:ID!){contextConfigAdd(context: $id, config: $input){id,checksum,type,value,writeOnly}}","variables":{"config":{"id":"SECRET_KEY","type":"ENVIRONMENT_VARIABLE","value":"dont-tell-anyone","writeOnly":true},"context":"babys-first-context"}}`,
		`{"data":{"contextConfigAdd":{"id":"context/babys-first-context/SECRET_KEY"}}}`,
		1,
	)

	e.posts(
		`{"query":"mutation($context:ID!$id:ID!){contextConfigDelete(context: $context, id: $id){id,checksum,type,value,writeOnly}}","variables":{"context":"babys-first-context","id":"SECRET_KEY"}}`,
		`{"data":{"contextConfigDelete":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_environment_variable" "context_variable" {
  context_id = "babys-first-context"
  name       = "SECRET_KEY"
  value      = "dont-tell-anyone"
}

data "spacelift_environment_variable" "context_variable" {
  context_id = "babys-first-context"
  name       = "SECRET_KEY"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "id", "context/babys-first-context/SECRET_KEY"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "checksum", "f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "context_id", "babys-first-context"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "name", "SECRET_KEY"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "value", ""),
				resource.TestCheckResourceAttr("spacelift_environment_variable.context_variable", "write_only", "true"),
				resource.TestCheckNoResourceAttr("spacelift_environment_variable.context_variable", "stack_id"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "id", "context/babys-first-context/SECRET_KEY"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "checksum", "f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "context_id", "babys-first-context"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "name", "SECRET_KEY"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "value", ""),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.context_variable", "write_only", "true"),
				resource.TestCheckNoResourceAttr("data.spacelift_environment_variable.context_variable", "stack_id"),
			),
		},
	})
}

func (e *EnvironmentVariableTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts(
		`{"query":"query($id:ID!$stack:ID!){stack(id: $stack){configElement(id: $id){id,checksum,type,value,writeOnly}}}","variables":{"id":"SECRET_KEY","stack":"babys-first-stack"}}`,
		`{"data":{"stack":{"configElement":{"id":"stack/babys-first-stack/SECRET_KEY","checksum":"f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1","type":"ENVIRONMENT_VARIABLE","value":null,"writeOnly":true}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($config:ConfigInput!$stack:ID!){stackConfigAdd(stack: $id, config: $input){id,checksum,type,value,writeOnly}}","variables":{"config":{"id":"SECRET_KEY","type":"ENVIRONMENT_VARIABLE","value":"dont-tell-anyone","writeOnly":true},"stack":"babys-first-stack"}}`,
		`{"data":{"stackConfigAdd":{"id":"stack/babys-first-stack/SECRET_KEY"}}}`,
		1,
	)

	e.posts(
		`{"query":"mutation($id:ID!$stack:ID!){stackConfigDelete(stack: $stack, id: $id){id,checksum,type,value,writeOnly}}","variables":{"id":"SECRET_KEY","stack":"babys-first-stack"}}`,
		`{"data":{"stackConfigDelete":{}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_environment_variable" "stack_variable" {
  stack_id = "babys-first-stack"
  name       = "SECRET_KEY"
  value      = "dont-tell-anyone"
}

data "spacelift_environment_variable" "stack_variable" {
  stack_id = "babys-first-stack"
  name       = "SECRET_KEY"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "id", "stack/babys-first-stack/SECRET_KEY"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "checksum", "f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "name", "SECRET_KEY"),
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "value", ""),
				resource.TestCheckResourceAttr("spacelift_environment_variable.stack_variable", "write_only", "true"),
				resource.TestCheckNoResourceAttr("spacelift_environment_variable.stack_variable", "context_id"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "id", "stack/babys-first-stack/SECRET_KEY"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "checksum", "f64b79a7d2493ddbb9aacaa69d6cca134ef190c8ed45a217cf640d93032af1c1"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "name", "SECRET_KEY"),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "value", ""),
				resource.TestCheckResourceAttr("data.spacelift_environment_variable.stack_variable", "write_only", "true"),
				resource.TestCheckNoResourceAttr("data.spacelift_environment_variable.stack_variable", "context_id"),
			),
		},
	})
}

func TestEnvironmentVariable(t *testing.T) {
	suite.Run(t, new(EnvironmentVariableTest))
}
