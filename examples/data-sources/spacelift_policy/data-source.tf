data "spacelift_policy" "policy" {
  policy_id = spacelift_policy.policy.id
}

output "policy_body" {
  value = data.spacelift_policy.policy.body
}
