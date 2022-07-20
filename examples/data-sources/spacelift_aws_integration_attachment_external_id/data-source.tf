# For a stack
data "spacelift_aws_integration_attachment_external_id" "my_stack" {
  integration_id = spacelift_aws_integration.this.id
  stack_id       = "my-stack-id"
  read           = true
  write          = true
}

# For a module
data "spacelift_aws_integration_attachment_external_id" "my_module" {
  integration_id = spacelift_aws_integration.this.id
  module_id      = "my-module-id"
  read           = true
  write          = true
}
