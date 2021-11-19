data "spacelift_current_stack" "this" {}

resource "spacelift_environment_variable" "core-kubeconfig" {
  stack_id = data.spacelift_current_stack.this.id
  name     = "CHUNKY"
  value    = "bacon"
}
