# Terraform stack using github.com as VCS
resource "spacelift_stack" "k8s-core" {
  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using Bitbucket Cloud as VCS
resource "spacelift_stack" "k8s-core-bitbucket-cloud" {
  bitbucket_cloud {
    namespace = "SPACELIFT" # The Bitbucket project containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using Bitbucket Data Center as VCS
resource "spacelift_stack" "k8s-core-bitbucket-datacenter" {
  bitbucket_datacenter {
    namespace = "SPACELIFT" # The Bitbucket project containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using GitHub Enterprise as VCS
resource "spacelift_stack" "k8s-core-github-enterprise" {
  github_enterprise {
    namespace = "spacelift" # The GitHub organization / user the repository belongs to
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using GitLab as VCS
resource "spacelift_stack" "k8s-core-gitlab" {
  gitlab {
    namespace = "spacelift" # The GitLab namespace containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Pulumi stack using github.com as VCS
resource "spacelift_stack" "k8s-core-pulumi" {
  pulumi {
    login_url  = "s3://pulumi-state-bucket"
    stack_name = "kubernetes-core-services"
  }

  autodeploy   = true
  branch       = "master"
  description  = "Shared cluster services (Datadog, Istio etc.)"
  name         = "Kubernetes core services"
  project_root = "/project"
  repository   = "core-infra"
  runner_image = "public.ecr.aws/t0p9w2l5/runner-pulumi-javascript:latest"
}
