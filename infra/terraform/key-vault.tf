resource "random_string" "kv_suffix" {
  length  = 4
  special = false
  upper   = false
}

resource "azurerm_key_vault" "hs_key_vault" {
  # Key Vault names must be globally unique across all Azure.
  name                       = "${local.project_name}-kv-${random_string.kv_suffix.result}"
  location                   = azurerm_resource_group.hs-rg.location
  resource_group_name        = azurerm_resource_group.hs-rg.name
  sku_name                   = "standard"
  soft_delete_retention_days = 30
  purge_protection_enabled   = true

  enable_rbac_authorization = true

  tenant_id = data.azurerm_client_config.current.tenant_id

  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
    ip_rules       = [var.external_ip_address]
  }
}

resource "azurerm_role_assignment" "kv_reader" {
  scope                = azurerm_key_vault.hs_key_vault.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.aks_workload_identity.principal_id
}


resource "azurerm_role_assignment" "kv_admin" {
  scope                = azurerm_key_vault.hs_key_vault.id
  role_definition_name = "Key Vault Administrator"
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "random_password" "psql_password" {
  length      = 20
  min_lower   = 1
  min_upper   = 1
  min_numeric = 1
  min_special = 1
  special     = false
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = true
  upper   = true
  lower   = true
  numeric = true
}

resource "azurerm_key_vault_secret" "jwt_secret" {
  name         = "jwt-secret"
  value        = random_password.jwt_secret.result
  key_vault_id = azurerm_key_vault.hs_key_vault.id

  depends_on = [azurerm_role_assignment.kv_admin]
}

resource "azurerm_key_vault_secret" "psql_password" {
  name         = "psql-password"
  value        = random_password.psql_password.result
  key_vault_id = azurerm_key_vault.hs_key_vault.id
  content_type = "text/plain"

  depends_on = [azurerm_role_assignment.kv_admin]
}

resource "azurerm_key_vault_secret" "psql_username" {
  name         = "psql-username"
  value        = local.psql_admin_username
  key_vault_id = azurerm_key_vault.hs_key_vault.id
  content_type = "text/plain"

  depends_on = [azurerm_role_assignment.kv_admin]
}


resource "azurerm_private_dns_zone" "kv_dns" {
  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = azurerm_resource_group.hs-rg.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "kv_dns_link" {
  name                  = "kv-dns-link"
  private_dns_zone_name = azurerm_private_dns_zone.kv_dns.name
  resource_group_name   = azurerm_resource_group.hs-rg.name
  virtual_network_id    = azurerm_virtual_network.hs-vn.id
  registration_enabled  = false
}


resource "azurerm_private_endpoint" "kv_pe" {
  name                = "kv-private-endpoint"
  location            = azurerm_resource_group.hs-rg.location
  resource_group_name = azurerm_resource_group.hs-rg.name
  subnet_id           = azurerm_subnet.hs-pl-subnet.id

  private_service_connection {
    name                           = "kv-private-connection"
    private_connection_resource_id = azurerm_key_vault.hs_key_vault.id
    subresource_names              = ["vault"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "kv-dns-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.kv_dns.id]
  }
}
