resource "spacelift_stack" "k8s-core" {
  // ...
}

// create the runs of a stack on a given schedule
resource "spacelift_scheduled_run" "k8s-core-apply" {
  stack_id = spacelift_stack.k8s-core.id

  name     = "apply-workdays"
  every    = ["0 7 * * 1-5"]
  timezone = "CET"
}

// run at a given timestamp (unix)
resource "spacelift_scheduled_run" "k8s-core-timestamp" {
  stack_id = spacelift_stack.k8s-core.id

  name = "one-off-apply"
  at   = "1663336895"
}

// run with custom runtime configuration
resource "spacelift_scheduled_run" "k8s-core-custom" {
  stack_id = spacelift_stack.k8s-core.id

  name     = "custom-terraform-apply"
  every    = ["0 21 * * 1-5"]
  timezone = "CET"
  runtime_config {
    environment {
      key   = SPACELIFT_DEBUG
      value = true
    }
  }
}
