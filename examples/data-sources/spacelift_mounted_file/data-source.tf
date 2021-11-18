# For a context
data "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
}

# For a module
data "spacelift_mounted_file" "module-kubeconfig" {
  module_id     = "k8s-module"
  relative_path = "kubeconfig"
}

# For a stack
data "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
}
