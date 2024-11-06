data "spacelift_tool_versions" "terraform" {
  tool = "TERRAFORM_FOSS"
}

output "terraform" {
  value = data.spacelift_tool_versions.terraform.versions
}

data "spacelift_tool_versions" "open_tofu" {
  tool = "OPEN_TOFU"
}

output "open_tofu" {
  value = data.spacelift_tool_versions.open_tofu.versions
}

data "spacelift_tool_versions" "kubectl" {
  tool = "KUBECTL"
}

output "kubectl" {
  value = data.spacelift_tool_versions.kubectl.versions
}

data "spacelift_tool_versions" "terragrunt" {
  tool = "TERRAGRUNT"
}

output "terragrunt" {
  value = data.spacelift_tool_versions.terragrunt.versions
}