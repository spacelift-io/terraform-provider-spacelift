variable "azure_tenant_id" {
  description = "Microsoft Entra tenant in which to install the Spacelift enterprise application."
  type        = string
}

variable "azure_subscription_id" {
  description = "Default Azure subscription for the Spacelift integration."
  type        = string
}

variable "name" {
  description = "Name of the Azure integration in Spacelift."
  type        = string
  default     = "Azure integration"
}

variable "space_id" {
  description = "Spacelift space in which to create the integration."
  type        = string
  default     = "root"
}

data "azuread_application_published_app_ids" "well_known" {}

data "azuread_service_principal" "microsoft_graph" {
  client_id = data.azuread_application_published_app_ids.well_known.result.MicrosoftGraph
}

resource "spacelift_azure_integration" "example" {
  name                    = var.name
  space_id                = var.space_id
  tenant_id               = var.azure_tenant_id
  default_subscription_id = var.azure_subscription_id
}

resource "azuread_service_principal" "spacelift" {
  client_id = spacelift_azure_integration.example.application_id
}

resource "azuread_service_principal_delegated_permission_grant" "admin_consent" {
  service_principal_object_id          = azuread_service_principal.spacelift.object_id
  resource_service_principal_object_id = data.azuread_service_principal.microsoft_graph.object_id
  claim_values                         = ["User.Read"]
}

resource "azurerm_role_assignment" "spacelift" {
  scope                = "/subscriptions/${var.azure_subscription_id}"
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.spacelift.object_id
  principal_type       = "ServicePrincipal"

  # The new service principal may not yet have replicated to Azure Resource Manager.
  skip_service_principal_aad_check = true
}

resource "spacelift_azure_integration_consent" "example" {
  integration_id = spacelift_azure_integration.example.id

  depends_on = [azuread_service_principal_delegated_permission_grant.admin_consent]
}
