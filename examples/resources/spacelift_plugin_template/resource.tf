resource "spacelift_plugin_template" "example" {
  name     = "opentofu-testing"
  manifest = file("./opentofu-testing-manifest.yaml")
}