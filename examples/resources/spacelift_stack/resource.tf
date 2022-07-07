# Terraform stack using github.com as VCS
resource "spacelift_stack" "k8s-cluster" {
  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using Bitbucket Cloud as VCS
resource "spacelift_stack" "k8s-cluster-bitbucket-cloud" {
  bitbucket_cloud {
    namespace = "SPACELIFT" # The Bitbucket project containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using Bitbucket Data Center as VCS
resource "spacelift_stack" "k8s-cluster-bitbucket-datacenter" {
  bitbucket_datacenter {
    namespace = "SPACELIFT" # The Bitbucket project containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using GitHub Enterprise as VCS
resource "spacelift_stack" "k8s-cluster-github-enterprise" {
  github_enterprise {
    namespace = "spacelift" # The GitHub organization / user the repository belongs to
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# Terraform stack using GitLab as VCS
resource "spacelift_stack" "k8s-cluster-gitlab" {
  gitlab {
    namespace = "spacelift" # The GitLab namespace containing the repository
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}

# CloudFormation stack using github.com as VCS
resource "spacelift_stack" "k8s-cluster-cloudformation" {
  cloudformation {
    entry_template_file = "main.yaml"
    region              = "eu-central-1"
    template_bucket     = "s3://bucket"
    stack_name          = "k8s-cluster"
  }

  autodeploy   = true
  branch       = "master"
  description  = "Provisions a Kubernetes cluster"
  name         = "Kubernetes Cluster"
  project_root = "cluster"
  repository   = "core-infra"
}

# Pulumi stack using github.com as VCS
resource "spacelift_stack" "k8s-cluster-pulumi" {
  pulumi {
    login_url  = "s3://pulumi-state-bucket"
    stack_name = "kubernetes-core-services"
  }

  autodeploy   = true
  branch       = "master"
  description  = "Provisions a Kubernetes cluster"
  name         = "Kubernetes Cluster"
  project_root = "cluster"
  repository   = "core-infra"
  runner_image = "public.ecr.aws/t0p9w2l5/runner-pulumi-javascript:latest"
}

# Kubernetes stack using github.com as VCS
resource "spacelift_stack" "k8s-core-kubernetes" {
  kubernetes {
    namespace = "core"
  }

  autodeploy   = true
  branch       = "master"
  description  = "Shared cluster services (Datadog, Istio etc.)"
  name         = "Kubernetes core services"
  project_root = "core-services"
  repository   = "core-infra"

  # You can use hooks to authenticate with your cluster
  before_init = ["aws eks update-kubeconfig --region us-east-2 --name k8s-cluster"]
}

# Ansible stack using github.com as VCS
resource "spacelift_stack" "ansible-stack" {
  ansible {
    playbook = "main.yml"
  }

  autodeploy   = true
  branch       = "master"
  description  = "Provisioning EC2 machines"
  name         = "Ansible EC2 playbooks"
  repository   = "ansible-playbooks"
  runner_image = "public.ecr.aws/spacelift/runner-ansible:latest"
}
