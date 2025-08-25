resource "random_string" "acr_suffix" {
  length  = 4
  special = false
  upper   = false
}

resource "azurerm_container_registry" "acr" {
  name                = "${local.project_name}${random_string.acr_suffix.result}"
  resource_group_name = azurerm_resource_group.hs-rg.name
  location            = azurerm_resource_group.hs-rg.location
  sku                 = "Basic"

  admin_enabled = false

  tags = {
    app = local.project_name
  }
}

# For production platform private link connection should be setup. It's available in premium tier though only
resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.aks.kubelet_identity[0].object_id
}
