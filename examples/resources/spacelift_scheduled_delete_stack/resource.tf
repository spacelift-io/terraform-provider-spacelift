resource "spacelift_stack" "k8s-core" {
  // ...
}

// at a given timestamp (unix)
resource "spacelift_scheduled_delete_stack" "k9s-core-delete" {
  stack_id = spacelift_stack.k8s-core.id

  at               = "1663336895"
  delete_resources = true
}
