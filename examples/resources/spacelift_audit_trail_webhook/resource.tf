resource "spacelift_audit_trail_webhook" "example" {
  endpoint = "https://example.com"
  enabled  = true
  secret   = "mysecretkey"
}

resource "spacelift_audit_trail_webhook" "write_only_example" {
  endpoint          = "https://example.com"
  enabled           = true
  secret_wo         = "somesupersecretkey"
  secret_wo_version = 1
}
