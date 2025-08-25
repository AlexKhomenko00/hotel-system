resource "azurerm_user_assigned_identity" "aks_workload_identity" {
  name                = "${local.project_name}-workload-identity"
  resource_group_name = azurerm_resource_group.hs-rg.name
  location            = azurerm_resource_group.hs-rg.location

  tags = {
    environment = "dev"
  }
}

resource "azurerm_federated_identity_credential" "aks_workload_federated_identity" {
  name                = "${local.project_name}-workload-federated-identity"
  resource_group_name = azurerm_resource_group.hs-rg.name
  audience            = ["api://AzureADTokenExchange"]
  issuer              = azurerm_kubernetes_cluster.aks.oidc_issuer_url
  parent_id           = azurerm_user_assigned_identity.aks_workload_identity.id
  subject             = "system:serviceaccount:default:workload-identity-sa"
}


resource "azurerm_kubernetes_cluster" "aks" {
  location            = azurerm_resource_group.hs-rg.location
  name                = "${local.project_name}-aks"
  resource_group_name = azurerm_resource_group.hs-rg.name
  dns_prefix          = local.project_name

  sku_tier = "Free"

  identity {
    type = "SystemAssigned"
  }

  default_node_pool {
    name           = "agentpool"
    vm_size        = "Standard_D2_v2"
    node_count     = var.node_count
    vnet_subnet_id = azurerm_subnet.hs-aks-subnet.id
  }

  linux_profile {
    admin_username = var.aks_username

    ssh_key {
      key_data = azapi_resource_action.ssh_public_key_gen.output.publicKey
    }
  }

  key_vault_secrets_provider {
    secret_rotation_enabled  = true
    secret_rotation_interval = "1h"
  }


  workload_identity_enabled = true
  oidc_issuer_enabled       = true

  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    pod_cidr            = "172.16.0.0/16" # Overlay network for pods
    service_cidr        = "10.124.0.0/16" # Kubernetes services
    dns_service_ip      = "10.124.0.10"   # CoreDNS service IP
  }

}


resource "azurerm_network_security_rule" "allow_https_from_aks" {
  name                        = "AllowHTTPSFromAKS"
  priority                    = 110
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = local.aks_subnet_cidr
  destination_address_prefix  = local.private_links_subnet_cidr
  resource_group_name         = azurerm_resource_group.hs-rg.name
  network_security_group_name = azurerm_network_security_group.hs-pl-sg.name
}
