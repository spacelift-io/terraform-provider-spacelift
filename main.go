package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var commit = "dev"
var version = "dev"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: spacelift.Provider(commit, version),
	})
}

provider "spacelift" {}

module "spacelift" {
  source  = "cloudposse/cloud-infrastructure-automation/spacelift"
  # Cloud Posse recommends pinning every module to a specific version
  # version     = "x.x.x"

  stack_config_path_template = var.stack_config_path_template
  components_path            = var.spacelift_component_path

  branch                = "main"
  repository            = var.git_repository
  commit_sha            = var.git_commit_sha
  spacelift_run_enabled = false
  runner_image          = var.runner_image
  worker_pool_id        = var.worker_pool_id
  autodeploy            = var.autodeploy
  manage_state          = false

  terraform_version     = var.terraform_version
  terraform_version_map = var.terraform_version_map

  imports_processing_enabled        = false
  stack_deps_processing_enabled     = false
  component_deps_processing_enabled = true

  policies_available     = var.policies_available
  policies_enabled       = var.policies_enabled
  policies_by_id_enabled = []
  policies_by_name_path  = format("%s/rego-policies", path.module)

  administrative_stack_drift_detection_enabled   = true
  administrative_stack_drift_detection_reconcile = true
  administrative_stack_drift_detection_schedule  = ["0 4 * * *"]

  drift_detection_enabled   = true
  drift_detection_reconcile = true
  drift_detection_schedule  = ["0 4 * * *"]

  aws_role_enabled                        = false
  aws_role_arn                            = null
  aws_role_external_id                    = null
  aws_role_generate_credentials_in_worker = false

  stack_destructor_enabled = false
}
