# Needed for generating the correct role ARNs
data "aws_caller_identity" "current" {}

locals {
  role_name = "my_role"
  role_arn  = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${local.role_name}"
}

# Create the AWS integration before creating your IAM role. The integration needs to exist
# in order to generate the external ID used for role assumption.
resource "spacelift_aws_integration" "this" {
  name = local.role_name

  # We need to set the ARN manually rather than referencing the role to avoid a circular dependency
  role_arn                       = local.role_arn
  generate_credentials_in_worker = false
}

data "spacelift_aws_integration_attachment_external_id" "my_stack" {
  integration_id = spacelift_aws_integration.this.id
  stack_id       = "my-stack-id"
  read           = true
  write          = true
}

data "spacelift_aws_integration_attachment_external_id" "my_module" {
  integration_id = spacelift_aws_integration.this.id
  module_id      = "my-module-id"
  read           = true
  write          = true
}

# Create the IAM role, using the `assume_role_policy_statement` from the data source.
resource "aws_iam_role" "this" {
  name = local.role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      jsondecode(data.spacelift_aws_integration_attachment_external_id.my_stack.assume_role_policy_statement),
      jsondecode(data.spacelift_aws_integration_attachment_external_id.my_module.assume_role_policy_statement),
    ]
  })
}

# For our example we're granting PowerUserAccess, but you can restrict this to whatever you need.
resource "aws_iam_role_policy_attachment" "this" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

# Attach the integration to any stacks or modules that need to use it
resource "spacelift_aws_integration_attachment" "my_stack" {
  integration_id = spacelift_aws_integration.this.id
  stack_id       = "my-stack-id"
  read           = true
  write          = true

  # The role needs to exist before we attach since we test role assumption during attachment.
  depends_on = [
    aws_iam_role.this
  ]
}

resource "spacelift_aws_integration_attachment" "my_module" {
  integration_id = spacelift_aws_integration.this.id
  module_id      = "my-module-id"
  read           = true
  write          = true

  # The role needs to exist before we attach since we test role assumption during attachment.
  depends_on = [
    aws_iam_role.this
  ]
}
