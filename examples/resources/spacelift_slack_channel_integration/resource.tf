resource "spacelift_slack_channel_integration" "slack_alerts" {
  integration_name = "Prod Alerts Channel"
  slack_channel_id = "C0123456789"

  access_rule {
    space_id = "prod-space"
    role     = "READ"
  }

  access_rule {
    space_id = "staging-space"
    role     = "WRITE"
  }
}
