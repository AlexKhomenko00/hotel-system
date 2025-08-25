variable "psqlf-sku" {
  type    = string
  default = "B_Standard_B1ms"
}

variable "psqlf-version" {
  type    = string
  default = "16"
}

variable "psqlf-storage-mb" {
  type    = number
  default = 32768
}

variable "az_psql_backup_retention_days" {
  type        = number
  description = "The number of days to retain backups for the PostgreSQL flexible server."
  default     = 7
}

variable "external_ip_address" {
  type        = string
  description = "External (Home maybe :) ) IP address to allow PostgreSQL access from"
}

variable "node_count" {
  type        = number
  description = "The initial quantity of nodes for the node pool."
  default     = 3
}

variable "aks_username" {
  type        = string
  description = "The admin username for the new cluster."
  default     = "azureadmin"
}
