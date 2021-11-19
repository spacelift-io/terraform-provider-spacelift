# For a stack
resource "spacelift_context_attachment" "attachment" {
  context_id = "prod-k8s-ie"
  stack_id   = "k8s-core"
  priority   = 0
}

# For a module
resource "spacelift_context_attachment" "attachment" {
  context_id = "prod-k8s-ie"
  module_id  = "k8s-module"
  priority   = 0
}
