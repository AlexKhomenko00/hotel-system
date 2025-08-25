terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.39.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~>3.0"
    }

    azapi = {
      source  = "azure/azapi"
      version = "~>1.5"
    }
  }
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

locals {
  project_name        = "hotel"
  psql_admin_username = "hsadmin"
}

resource "azurerm_resource_group" "hs-rg" {
  name     = "hs-resources"
  location = "West Europe"

  tags = {
    environment = "dev"
  }
}
