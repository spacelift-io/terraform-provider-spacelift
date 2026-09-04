# Explicit module name and provider:
resource "spacelift_module" "k8s-module" {
  name               = "k8s-module"
  terraform_provider = "aws"
  branch             = "master"
  description        = "Infra terraform module"
  repository         = "terraform-super-module"
}

# Terraform module using the installed account-level Origin VCS integration.
resource "spacelift_module" "origin-module" {
  origin {}

  branch     = "main"
  name       = "origin-module"
  repository = "terraform-origin-module"
}

# Unspecified module name and provider (repository naming scheme terraform-${provider}-${name})
resource "spacelift_module" "example-module" {
  branch       = "master"
  description  = "Example terraform module"
  repository   = "terraform-aws-example"
  project_root = "example"
}
