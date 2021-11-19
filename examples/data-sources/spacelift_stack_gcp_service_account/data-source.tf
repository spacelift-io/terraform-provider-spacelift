# For a Module
data "spacelift_stack_gcp_service_account" "k8s-module" {
  module_id = "k8s-module"
}

# For a Stack
data "spacelift_stack_gcp_service_account" "k8s-core" {
  stack_id = "k8s-core"
}
