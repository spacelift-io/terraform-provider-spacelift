# With a stack.
data "spacelift_context_attachment" "apps-k8s-ie" {
  context_id = "prod-k8s-ie"
  stack_id   = "apps-cluster"
}

# With a module.
data "spacelift_context_attachment" "kafka-k8s-ie" {
  context_id = "prod-k8s-ie"
  module_id  = "terraform-aws-kafka"
}
