resource "spacelift_stack" "core-infra-production" {
  name       = "Core Infrastructure (production)"
  branch     = "master"
  repository = "core-infra"
}

resource "spacelift_drift_detection" "core-infra-production-drift-detection" {
  reconcile = true
  stack_id  = spacelift_stack.core-infra-production.id
  schedule  = ["*/15 * * * *"] # Every 15 minutes
}
