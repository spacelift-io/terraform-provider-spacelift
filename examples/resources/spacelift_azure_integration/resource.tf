resource "spacelift_azure_integration" "example" {
  name                    = "Example integration"
  tenant_id               = "tenant-id"
  default_subscription_id = "default-subscription-id"
  labels                  = ["one", "two"]
}
