locals {
  vnet_cidr                 = "10.123.0.0/16"
  aks_subnet_cidr           = "10.123.1.0/24"
  private_links_subnet_cidr = "10.123.2.0/24"
}

resource "azurerm_virtual_network" "hs-vn" {
  name                = "hs-network"
  location            = azurerm_resource_group.hs-rg.location
  resource_group_name = azurerm_resource_group.hs-rg.name
  address_space       = [local.vnet_cidr]

  tags = {
    environment = "dev"
  }
}

resource "azurerm_subnet" "hs-aks-subnet" {
  name                 = "hs-aks-subnet"
  resource_group_name  = azurerm_resource_group.hs-rg.name
  virtual_network_name = azurerm_virtual_network.hs-vn.name

  address_prefixes = [local.aks_subnet_cidr]
}


resource "azurerm_network_security_group" "hs-aks-sg" {
  name                = "hs-aks-sg"
  location            = azurerm_resource_group.hs-rg.location
  resource_group_name = azurerm_resource_group.hs-rg.name

  tags = {
    environment = "dev"
  }
}

resource "azurerm_subnet_network_security_group_association" "hs-aks-subnet-sga" {
  subnet_id                 = azurerm_subnet.hs-aks-subnet.id
  network_security_group_id = azurerm_network_security_group.hs-aks-sg.id
}

resource "azurerm_subnet" "hs-pl-subnet" {
  name                 = "hs-private-links-subnet"
  resource_group_name  = azurerm_resource_group.hs-rg.name
  virtual_network_name = azurerm_virtual_network.hs-vn.name

  address_prefixes = [local.private_links_subnet_cidr]
}


resource "azurerm_network_security_group" "hs-pl-sg" {
  name                = "hs-pl-sg"
  location            = azurerm_resource_group.hs-rg.location
  resource_group_name = azurerm_resource_group.hs-rg.name

  tags = {
    environment = "dev"
  }
}

resource "azurerm_subnet_network_security_group_association" "hs-pl-subnet-sga" {
  subnet_id                 = azurerm_subnet.hs-pl-subnet.id
  network_security_group_id = azurerm_network_security_group.hs-pl-sg.id
}

# Allow PostgreSQL traffic from AKS subnet to private links subnet
resource "azurerm_network_security_rule" "allow_psql_from_aks" {
  name                        = "AllowPostgreSQLFromAKS"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "5432"
  source_address_prefix       = local.aks_subnet_cidr
  destination_address_prefix  = local.private_links_subnet_cidr
  resource_group_name         = azurerm_resource_group.hs-rg.name
  network_security_group_name = azurerm_network_security_group.hs-pl-sg.name
}

# Allow outbound traffic for private endpoint communication
resource "azurerm_network_security_rule" "allow_outbound_https" {
  name                        = "AllowOutboundHTTPS"
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = local.private_links_subnet_cidr
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.hs-rg.name
  network_security_group_name = azurerm_network_security_group.hs-pl-sg.name
}




