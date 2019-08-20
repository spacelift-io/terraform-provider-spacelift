package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type ContextTest struct {
	ResourceTest
}

func (e *ContextTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts( // Mocking out the context creation.
		`{"query":"mutation($description:String$name:String!){contextCreate(name: $name, description: $description){id,description,name}}","variables":{"description":"This is an example description","name":"Baby's first context"}}`,
		`{"data":{"contextCreate":{"id":"babys-first-context"}}}`,
		1,
	)

	e.posts( // Mocking out the context query.
		`{"query":"query($id:ID!){context(id: $id){id,description,name}}","variables":{"id":"babys-first-context"}}`,
		`{"data":{"context":{"id":"babys-first-context","description":"This is an example description","name":"Baby's first context"}}}`,
		6,
	)

	e.posts( // Mocking out the context deletion.
		`{"query":"mutation($id:String!){contextDelete(id: $id){id,administrative,awsAssumedRoleARN,branch,description,name,readersSlug,repo,terraformVersion,writersSlug}}","variables":{"id":"babys-first-context"}}`,
		`{"data":{"contextDelete":{"id":"babys-first-context"}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{{
		Config: `
resource "spacelift_context" "context" {
  description = "This is an example description"
  name        = "Baby's first context"
}

data "spacelift_context" "context" {
	context_id = spacelift_context.context.id
}
`,
		Check: resource.ComposeTestCheckFunc(
			// Test resource.
			resource.TestCheckResourceAttr("spacelift_context.context", "id", "babys-first-context"),
			resource.TestCheckResourceAttr("spacelift_context.context", "name", "Baby's first context"),
			resource.TestCheckResourceAttr("spacelift_context.context", "description", "This is an example description"),

			// Test data.
			resource.TestCheckResourceAttr("data.spacelift_context.context", "id", "babys-first-context"),
			resource.TestCheckResourceAttr("data.spacelift_context.context", "name", "Baby's first context"),
			resource.TestCheckResourceAttr("data.spacelift_context.context", "description", "This is an example description"),
		),
	}})
}

func TestContext(t *testing.T) {
	suite.Run(t, new(ContextTest))
}
