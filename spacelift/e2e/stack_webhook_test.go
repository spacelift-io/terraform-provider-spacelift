package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type StackWebhookTest struct {
	ResourceTest
}

func (e *StackWebhookTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($input:WebhooksIntegrationInput!$stack:ID!){webhooksIntegrationCreate(stack: $stack, input: $input){id,enabled}}","variables":{"input":{"enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"},"stack":"babys-first-stack"}}`,
		`{"data":{"webhooksIntegrationCreate":{"id":"babys-first-webhook", "enabled": true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,deleted,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"integrations":{"webhooks":[{"id":"babys-first-webhook","deleted":false,"enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"}]}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($stack:ID!$webhook:ID!){webhooksIntegrationDelete(stack: $stack, id: $webhook){id}}","variables":{"stack":"babys-first-stack","webhook":"babys-first-webhook"}}`,
		`{"data":{"webhooksIntegrationDelete":{"id":"babys-first-webhook"}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_stack_webhook" "webhook" {
  stack_id = "babys-first-stack"

  endpoint = "localtest.me/amazin"
  secret = "test_secret2"
}

data "spacelift_stack_webhook" "webhook" {
  stack_id = "babys-first-stack"
  webhook_id = "babys-first-webhook"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "deleted", "false"),
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("spacelift_stack_webhook.webhook", "secret", "test_secret2"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "deleted", "false"),
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("data.spacelift_stack_webhook.webhook", "secret", "test_secret2"),
			),
		},
	})
}

func TestStackWebhook(t *testing.T) {
	suite.Run(t, new(StackWebhookTest))
}
