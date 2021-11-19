resource "spacelift_webhook" "webhook" {
  endpoint = "https://example.com/webhooks"
  stack_id = "k8s-core"
}
