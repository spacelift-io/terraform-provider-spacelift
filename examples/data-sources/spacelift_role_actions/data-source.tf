data "spacelift_role_actions" "actions" {}

output "possible_role_actions" {
  value = data.spacelift_role_actions.actions.actions
}
