resource "spacelift_stack" "k8s-core" {
  // ...
}

resource "spacelift_environment_variable" "credentials" {
  // ...
}

resource "spacelift_stack_destructor" "k8s-core" {
  depends_on = [
    spacelift_environment_variable.credentials,
  ]

  stack_id = spacelift_stack.k8s-core.id
}
