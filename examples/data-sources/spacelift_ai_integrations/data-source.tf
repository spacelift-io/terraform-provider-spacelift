data "spacelift_ai_integrations" "all" {}

# Only the integrations carrying every one of these labels.
data "spacelift_ai_integrations" "tagged" {
  labels = ["ai"]
}
