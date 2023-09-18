resource "spacelift_stack" "k8s-core" {
  // ...
}

// create the resources of a stack on a given schedule
resource "spacelift_scheduled_task" "k8s-core-create" {
  stack_id = spacelift_stack.k8s-core.id

  command  = "terraform apply -auto-approve"
  every    = ["0 7 * * 1-5"]
  timezone = "CET"
}

// destroy the resources of a stack on a given schedule
resource "spacelift_scheduled_task" "k8s-core-destroy" {
  stack_id = spacelift_stack.k8s-core.id

  command  = "terraform destroy -auto-approve"
  every    = ["0 21 * * 1-5"]
  timezone = "CET"
}

// at a given timestamp (unix)
resource "spacelift_scheduled_task" "k8s-core-destroy" {
  stack_id = spacelift_stack.k8s-core.id

  command = "terraform destroy -auto-approve"
  at      = "1663336895"
}