# For a Module
data "spacelift_stack_aws_role" "k8s-module" {
  module_id = "k8s-module"
}

# For a Stack
data "spacelift_stack_aws_role" "k8s-core" {
  stack_id = "k8s-core"
}
