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
