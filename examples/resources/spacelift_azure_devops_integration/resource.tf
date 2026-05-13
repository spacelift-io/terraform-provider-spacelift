resource "spacelift_azure_devops_integration" "example" {
  name                  = "Azure DevOps integration (public)"
  space_id              = "root"
  organization_url      = "https://dev.azure.com/my-organization"
  personal_access_token = "azure-devops-token"
}

resource "spacelift_azure_devops_integration" "private-example" {
  name                  = "Azure DevOps integration (private)"
  is_default            = true
  organization_url      = "https://dev.azure.com/my-organization"
  user_facing_host      = "private://my-vcs-agent-pool/my-organization"
  personal_access_token = "azure-devops-token"

  accessible_projects = ["Project One", "Project Two"]
}

resource "spacelift_azure_devops_integration" "example-write-only" {
  name                             = "Azure DevOps integration (public)"
  space_id                         = "root"
  organization_url                 = "https://dev.azure.com/my-organization"
  personal_access_token_wo         = "azure-devops-token"
  personal_access_token_wo_version = 1
}