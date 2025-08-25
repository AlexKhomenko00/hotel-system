resource "azurerm_postgresql_flexible_server" "hs-psqlf" {
  name                = "psqlf-${local.project_name}"
  resource_group_name = azurerm_resource_group.hs-rg.name
  location            = azurerm_resource_group.hs-rg.location

  depends_on = [azurerm_resource_group.hs-rg]

  sku_name              = var.psqlf-sku
  version               = var.psqlf-version
  storage_mb            = var.psqlf-storage-mb
  backup_retention_days = var.az_psql_backup_retention_days

  # To access from home ip. Normally disabled on production instance can be accessed through  smth like hub or bastion.
  public_network_access_enabled = true

  administrator_login    = local.psql_admin_username
  administrator_password = random_password.psql_password.result

  lifecycle {
    ignore_changes = [zone]
  }
}

resource "azurerm_private_dns_zone" "hs-psqlf-dns" {
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = azurerm_resource_group.hs-rg.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "hs-psqlf-dns-link" {
  name                  = "psql-vnet-link"
  resource_group_name   = azurerm_resource_group.hs-rg.name
  private_dns_zone_name = azurerm_private_dns_zone.hs-psqlf-dns.name
  virtual_network_id    = azurerm_virtual_network.hs-vn.id
  registration_enabled  = false
}

resource "azurerm_private_endpoint" "hs-psglf-pe" {
  name                = "psql-private-endpoint"
  location            = azurerm_resource_group.hs-rg.location
  resource_group_name = azurerm_resource_group.hs-rg.name
  subnet_id           = azurerm_subnet.hs-pl-subnet.id

  private_service_connection {
    name                           = "psql-private-connection"
    private_connection_resource_id = azurerm_postgresql_flexible_server.hs-psqlf.id
    subresource_names              = ["postgresqlServer"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "psql-dns-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.hs-psqlf-dns.id]
  }
}

resource "azurerm_postgresql_flexible_server_firewall_rule" "home_access" {
  name             = "AllowHomeIP"
  server_id        = azurerm_postgresql_flexible_server.hs-psqlf.id
  start_ip_address = var.external_ip_address
  end_ip_address   = var.external_ip_address
}

