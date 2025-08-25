
output "acr_login_server" {
  value = azurerm_container_registry.acr.login_server
}

output "acr_name" {
  value = azurerm_container_registry.acr.name
}

output "aks_cluster_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

output "workload_identity_client_id" {
  value       = azurerm_user_assigned_identity.aks_workload_identity.client_id
  description = "Use this in your K8s service account annotation"
}

output "key_vault_name" {
  value = azurerm_key_vault.hs_key_vault.name
}

output "kube_config" {
  value     = azurerm_kubernetes_cluster.aks.kube_config_raw
  sensitive = true
}

output "user_assigned_identity_client_id" {
  value = azurerm_user_assigned_identity.aks_workload_identity.client_id
}

output "tenant_id" {
  value = data.azurerm_client_config.current.tenant_id
}

output "postgresql_server_fqdn" {
  value = azurerm_postgresql_flexible_server.hs-psqlf.fqdn
}

output "postgresql_server_port" {
  value       = "5432"
  description = "PostgreSQL server port"
}

output "postgresql_database_name" {
  # https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/quickstart-create-server?tabs=cli-create-flexible%2Ccli-create-get-connection%2Cportal-delete-resources#databases-available-in-an-azure-database-for-postgresql-flexible-server-instance
  value       = "postgres"
  description = "PostgreSQL database name"
}

output "resource_group_name" {
  value = azurerm_resource_group.hs-rg.name
}


