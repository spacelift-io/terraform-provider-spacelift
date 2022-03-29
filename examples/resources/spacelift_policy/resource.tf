resource "spacelift_policy" "no-weekend-deploys" {
  name = "Let's not deploy any changes over the weekend"
  body = file("${path.module}/policies/no-weekend-deploys.rego")
  type = "PLAN"
}

resource "spacelift_stack" "core-infra-production" {
  name       = "Core Infrastructure (production)"
  branch     = "master"
  repository = "core-infra"
}

resource "spacelift_policy_attachment" "no-weekend-deploys" {
  policy_id = spacelift_policy.no-weekend-deploys.id
  stack_id  = spacelift_stack.core-infra-production.id
}
