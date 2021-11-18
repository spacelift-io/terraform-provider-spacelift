# Assuming the role in Spacelift
resource "aws_iam_role" "spacelift" {
  name = "spacelift"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [jsondecode(spacelift_stack.k8s-core.aws_assume_role_policy_statement)]
  })
}

resource "aws_iam_role_policy_attachment" "spacelift" {
  role       = aws_iam_role.spacelift.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

# Role assumed by a Stack
resource "spacelift_stack_aws_role" "spacelift-stack" {
  stack_id = "k8s-core"
  role_arn = aws_iam_role.spacelift.arn
}

# Role assumed by a Module
resource "spacelift_stack_aws_role" "spacelift-module" {
  module_id = "k8s-core"
  role_arn  = aws_iam_role.spacelift.arn
}

# Assuming the role in the private worker, for a stack.
resource "spacelift_stack_aws_role" "k8s-core" {
  stack_id                       = "k8s-core"
  role_arn                       = "arn:aws:iam::123456789012:custom/role"
  generate_credentials_in_worker = true
}

# Assuming the role in the private worker, for a module.
resource "spacelift_stack_aws_role" "k8s-core" {
  module_id                      = "k8s-core"
  role_arn                       = "arn:aws:iam::123456789012:custom/role"
  generate_credentials_in_worker = true
}
