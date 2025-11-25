data "spacelift_plugin" "opentofu_tracing" {
  plugin_id = "opentofu-tracing"
}

resource "spacelift_stack" "example" {
  name   = "example-stack-with-opentofu-tracing-plugin"
  labels = [data.spacelift_plugin.opentofu_tracing.stack_label]
}