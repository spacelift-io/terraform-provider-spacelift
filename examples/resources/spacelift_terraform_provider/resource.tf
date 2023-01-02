resource "spacelift_terraform_provider" "datadog" {
  type     = "datadog"
  space_id = "root"

  description = "Our fork of the Datadog provider"
  labels      = ["fork"]
  public      = false
}
