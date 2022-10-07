data "spacelift_current_space" "this" {}

resource "spacelift_context" "prod-k8s-ie" {
  description = "Configuration details for the compute cluster in 🇮🇪"
  name        = "Production cluster (Ireland)"
  space_id    = data.spacelift_current_space.this.id
}