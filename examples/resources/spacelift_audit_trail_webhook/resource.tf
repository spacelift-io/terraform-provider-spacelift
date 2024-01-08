resource "spacelift_audit_trail_webhook" "example" {
  endpoint = "https://example.com"
  enabled  = true
  secret   = "mysecretkey"
}
