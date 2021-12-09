# For a stack
resource "spacelift_azure_integration_attachment" "readonly" {
  integration_id  = spacelift_azure_integration.example.id
  stack_id        = spacelift_stack.example.id
  write           = false
  subscription_id = "subscription_id"
}

# For a module
resource "spacelift_azure_integration_attachment" "writeonly" {
  integration_id  = spacelift_azure_integration.example.id
  stack_id        = spacelift_module.example.id
  read            = false
  subscription_id = "subscription_id"
}
