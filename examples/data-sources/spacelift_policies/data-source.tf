# For all policies
data "spacelift_policies" "all" {}

# Policies with a matching type & all specified labels
data "spacelift_policies" "plan_autoattach" {
  type   = "PLAN"
  labels = ["autoattach"]
}

output "policy_ids" {
  value = data.spacelift_policies.this.policies[*].id
}
