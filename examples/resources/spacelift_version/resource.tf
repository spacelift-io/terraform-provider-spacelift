resource "spacelift_module" "k8s-module" {
  name               = "k8s-module"
  terraform_provider = "aws"
  administrative     = true
  branch             = "master"
  description        = "Infra terraform module"
  repository         = "terraform-super-module"
}

resource "spacelift_version" "k8s-module" {
  module_id  = spacelift_module.k8s-module.id
  commit_sha = "abc123def456789"
  keepers = {
    vpc_config = data.aws_eks_cluster.k8s-module.vpc_config[0].security_group_ids
  }
  version_number = "1.2.3"
}