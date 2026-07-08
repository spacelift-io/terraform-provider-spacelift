resource "spacelift_intent_project" "sandbox" {
  name             = "Ephemeral sandbox"
  space_id         = "root"
  description      = "Sandbox environment, cleaned up automatically after three days"
  labels           = ["intent", "sandbox"]
  ttl              = "72h"
  on_expiry_action = "DELETE"
}