# For a context
data "spacelift_environment_variable" "ireland-kubeconfig" {
  context_id = "prod-k8s-ie"
  name       = "KUBECONFIG"
}

# For a module
data "spacelift_environment_variable" "module-kubeconfig" {
  module_id = "k8s-module"
  name      = "KUBECONFIG"
}

# For a stack
data "spacelift_environment_variable" "core-kubeconfig" {
  stack_id = "k8s-core"
  name     = "KUBECONFIG"
}
