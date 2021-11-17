# With a module
data "spacelift_aws_role" "k8s-module" {
  module_id = "k8s-module"
}

# With a stack
data "spacelift_aws_role" "k8s-core" {
  stack_id = "k8s-core"
}
