# For a context
resource "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}

# For a module
resource "spacelift_mounted_file" "module-kubeconfig" {
  module_id     = "k8s-module"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}

# For a stack
resource "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}

# For a stack with file_mode
resource "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
  file_mode     = "755"
}
