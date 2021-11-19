resource "spacelift_stack" "this" {
  name       = "Test stack"
  repository = "test"
  branch     = "main"
}

resource "spacelift_run" "this" {
  stack_id = spacelift_stack.this.id

  keepers = {
    branch = spacelift_stack.this.branch
  }
}
