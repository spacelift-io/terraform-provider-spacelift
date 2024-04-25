# Terraform stack using github.com as VCS
resource "spacelift_stack" "k8s-cluster" {
  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Provisions a Kubernetes cluster"
  name              = "Kubernetes Cluster"
  project_root      = "cluster"
  repository        = "core-infra"
  terraform_version = "1.3.0"
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
  terraform_version = "1.3.0"
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
  terraform_version = "1.3.0"
}

# Terraform stack using a GitHub Custom Application. See the following page for more info: https://docs.spacelift.io/integrations/source-control/github#setting-up-the-custom-application
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
  terraform_version = "1.3.0"
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
  terraform_version = "1.3.0"
}

# Terraform stack using github.com as VCS and enabling smart sanitization
resource "spacelift_stack" "k8s-cluster" {
  administrative               = true
  autodeploy                   = true
  branch                       = "master"
  description                  = "Provisions a Kubernetes cluster"
  name                         = "Kubernetes Cluster"
  project_root                 = "cluster"
  repository                   = "core-infra"
  terraform_version            = "1.3.0"
  terraform_smart_sanitization = true
}

# Terraform stack using github.com as VCS and enabling external state access
resource "spacelift_stack" "k8s-cluster" {
  administrative                  = true
  autodeploy                      = true
  branch                          = "master"
  description                     = "Provisions a Kubernetes cluster"
  name                            = "Kubernetes Cluster"
  project_root                    = "cluster"
  repository                      = "core-infra"
  terraform_version               = "1.3.0"
  terraform_external_state_access = true
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
    namespace       = "core"
    kubectl_version = "1.26.1" # Optional kubectl version
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

# Terragrunt stack using github.com as VCS
resource "spacelift_stack" "terragrunt-stack" {
  terragrunt {
    terraform_version      = "1.6.2"
    terragrunt_version     = "0.55.15"
    use_run_all            = false
    use_smart_sanitization = true
    tool                   = "OPEN_TOFU"
  }

  autodeploy   = true
  branch       = "main"
  name         = "Terragrunt stack example"
  description  = "Deploys infra using Terragrunt"
  repository   = "terragrunt-stacks"
  project_root = "path/to/terragrunt_hcl"
}
