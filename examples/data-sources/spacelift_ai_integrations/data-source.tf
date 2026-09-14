data "spacelift_ai_integrations" "all" {}

# Only the integrations carrying every one of these labels.
data "spacelift_ai_integrations" "tagged" {
  labels = ["ai"]
}

# The filters combine, and each one is only applied when it is set: leaving
# spacelift_provided out returns both kinds, false returns only your own.
data "spacelift_ai_integrations" "own_anthropic" {
  ai_provider        = "Anthropic"
  spacelift_provided = false
}
