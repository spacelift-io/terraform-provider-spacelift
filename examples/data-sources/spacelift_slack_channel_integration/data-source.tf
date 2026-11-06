data "spacelift_slack_channel_integration" "prod_alerts" {
  integration_id = "some-integration-id"
}

output "integration_name" {
  value = data.spacelift_slack_channel_integration.prod_alerts.integration_name
}

output "slack_channel_id" {
  value = data.spacelift_slack_channel_integration.prod_alerts.slack_channel_id
}
