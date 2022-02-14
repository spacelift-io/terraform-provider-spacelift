# Explicit module name and provider:
resource "spacelift_module" "k8s-module" {
  name               = "k8s-module"
  terraform_provider = "aws"
  administrative     = true
  branch             = "master"
  description        = "Infra terraform module"
  repository         = "terraform-super-module"
}

# Unspecified module name and provider (repository naming scheme terraform-${provider}-${name})
resource "spacelift_module" "example-module" {
  administrative     = true
  branch             = "master"
  description        = "Example terraform module"
  repository         = "terraform-aws-example"
  project_root       = "example"
}
