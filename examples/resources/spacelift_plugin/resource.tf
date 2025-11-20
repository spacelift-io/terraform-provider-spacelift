data "spacelift_plugin_template" "opentofu_tracing" {
  plugin_template_id = "opentofu-tracing"
}

resource "spacelift_plugin" "opentofu_tracing" {
  plugin_template_id = data.spacelift_plugin_template.opentofu_tracing.id
  name               = "OpenTofu Tracing Plugin"
  stack_label        = "tracing"
  parameters = {
    output_file = "tracing_output.md"
  }
}