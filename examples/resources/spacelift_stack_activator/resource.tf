resource "spacelift_stack" "app" {
  branch     = "master"
  name       = "Application stack"
  repository = "app"
}

resource "spacelift_stack_activator" "test" {
  enabled  = true
  stack_id = spacelift_stack.app.id
}
