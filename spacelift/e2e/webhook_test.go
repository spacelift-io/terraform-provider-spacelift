package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type WebhookTest struct {
	ResourceTest
}

func (e *WebhookTest) TestLifecycle_Module() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($input:WebhooksIntegrationInput!$stack:ID!){webhooksIntegrationCreate(stack: $stack, input: $input){id,enabled}}","variables":{"input":{"enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"},"stack":"babys-first-module"}}`,
		`{"data":{"webhooksIntegrationCreate":{"id":"babys-first-webhook", "enabled": true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){module(id: $id){id,administrative,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,enabled,endpoint,secret}},labels,namespace,provider,repository,workerPool{id}}}","variables":{"id":"babys-first-module"}}`,
		`{"data":{"module":{"integrations":{"webhooks":[{"id":"babys-first-webhook","enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"}]}}}}`,
		7,
	)

	e.posts(
		`{"query":"mutation($stack:ID!$webhook:ID!){webhooksIntegrationDelete(stack: $stack, id: $webhook){id}}","variables":{"stack":"babys-first-module","webhook":"babys-first-webhook"}}`,
		`{"data":{"webhooksIntegrationDelete":{"id":"babys-first-webhook"}}}`,
		1,
	)

	e.testsResource([]resource.TestStep{
		{
			Config: `
resource "spacelift_webhook" "webhook" {
  module_id = "babys-first-module"

  endpoint = "localtest.me/amazin"
  secret = "test_secret2"
}

data "spacelift_webhook" "webhook" {
  module_id = "babys-first-module"
  webhook_id = "babys-first-webhook"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "secret", "test_secret2"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "module_id", "babys-first-module"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "secret", "test_secret2"),
			),
		},
	})
}

func (e *WebhookTest) TestLifecycle_Stack() {
	defer gock.Off()

	e.posts(
		`{"query":"mutation($input:WebhooksIntegrationInput!$stack:ID!){webhooksIntegrationCreate(stack: $stack, input: $input){id,enabled}}","variables":{"input":{"enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"},"stack":"babys-first-stack"}}`,
		`{"data":{"webhooksIntegrationCreate":{"id":"babys-first-webhook", "enabled": true}}}`,
		1,
	)

	e.posts(
		`{"query":"query($id:ID!){stack(id: $id){id,administrative,autodeploy,branch,description,integrations{aws{assumedRoleArn,assumeRolePolicyStatement},gcp{serviceAccountEmail,tokenScopes},webhooks{id,enabled,endpoint,secret}},labels,managesStateFile,name,namespace,projectRoot,provider,repository,terraformVersion,workerPool{id}}}","variables":{"id":"babys-first-stack"}}`,
		`{"data":{"stack":{"integrations":{"webhooks":[{"id":"babys-first-webhook","enabled":true,"endpoint":"localtest.me/amazin","secret":"test_secret2"}]}}}}`,
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
resource "spacelift_webhook" "webhook" {
  stack_id = "babys-first-stack"

  endpoint = "localtest.me/amazin"
  secret = "test_secret2"
}

data "spacelift_webhook" "webhook" {
  stack_id = "babys-first-stack"
  webhook_id = "babys-first-webhook"
}
`,
			Check: resource.ComposeTestCheckFunc(
				// Test resource.
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("spacelift_webhook.webhook", "secret", "test_secret2"),

				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "stack_id", "babys-first-stack"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "id", "babys-first-webhook"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "enabled", "true"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "endpoint", "localtest.me/amazin"),
				resource.TestCheckResourceAttr("data.spacelift_webhook.webhook", "secret", "test_secret2"),
			),
		},
	})
}

func TestWebhook(t *testing.T) {
	suite.Run(t, new(WebhookTest))
}
