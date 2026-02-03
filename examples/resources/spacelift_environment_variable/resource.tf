# For a context
resource "spacelift_environment_variable" "ireland-kubeconfig" {
  context_id  = "prod-k8s-ie"
  name        = "KUBECONFIG"
  value       = "/project/spacelift/kubeconfig"
  write_only  = false
  description = "Kubeconfig for Ireland Kubernetes cluster"
}

# For a module
resource "spacelift_environment_variable" "module-kubeconfig" {
  module_id   = "k8s-module"
  name        = "KUBECONFIG"
  value       = "/project/spacelift/kubeconfig"
  write_only  = false
  description = "Kubeconfig for the module"
}

# For a stack
resource "spacelift_environment_variable" "core-kubeconfig" {
  stack_id    = "k8s-core"
  name        = "KUBECONFIG"
  value       = "/project/spacelift/kubeconfig"
  write_only  = false
  description = "Kubeconfig for the core stack"
}

# For a secret variable using write-only attributes (recommended for secrets)
# The value is never stored in Terraform state. Requires Terraform 1.11+ or OpenTofu 1.11+.
resource "spacelift_environment_variable" "api-token" {
  context_id       = "prod-k8s-ie"
  name             = "API_TOKEN"
  value_wo         = var.api_token
  value_wo_version = 1 # Increment to update the value
  write_only       = true
  description      = "API token for external service"
}
