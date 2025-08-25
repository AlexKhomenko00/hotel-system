#!/bin/bash

# Azure K8s Deployment Script
# This script reads Terraform outputs and deploys to Azure Kubernetes Service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/terraform"
K8S_OVERLAY_DIR="$SCRIPT_DIR/k8s/overlays/azure"

echo "🚀 Starting Azure Kubernetes deployment..."

# Check if terraform directory exists and has state
if [ ! -d "$TERRAFORM_DIR" ]; then
    echo "❌ Terraform directory not found: $TERRAFORM_DIR"
    exit 1
fi

if [ ! -f "$TERRAFORM_DIR/terraform.tfstate" ]; then
    echo "❌ Terraform state not found. Please run 'terraform apply' first."
    exit 1
fi

# Change to terraform directory to get outputs
cd "$TERRAFORM_DIR"

echo "📄 Reading Terraform outputs..."

# Get Terraform outputs
ACR_LOGIN_SERVER=$(terraform output -raw acr_login_server 2>/dev/null || echo "")
WORKLOAD_IDENTITY_CLIENT_ID=$(terraform output -raw workload_identity_client_id 2>/dev/null || echo "")
KEY_VAULT_NAME=$(terraform output -raw key_vault_name 2>/dev/null || echo "")
TENANT_ID=$(terraform output -raw tenant_id 2>/dev/null || echo "")
CLUSTER_NAME=$(terraform output -raw aks_cluster_name 2>/dev/null || echo "")
DB_HOST=$(terraform output -raw postgresql_server_fqdn 2>/dev/null || echo "")
DB_PORT=$(terraform output -raw postgresql_server_port 2>/dev/null || echo "5432")
DB_DATABASE=$(terraform output -raw postgresql_database_name 2>/dev/null || echo "postgres")

# Validate required outputs
if [ -z "$ACR_LOGIN_SERVER" ] || [ -z "$WORKLOAD_IDENTITY_CLIENT_ID" ] || [ -z "$KEY_VAULT_NAME" ] || [ -z "$TENANT_ID" ]; then
    echo "❌ Missing required Terraform outputs. Please ensure all resources are properly deployed."
    echo "   ACR_LOGIN_SERVER: $ACR_LOGIN_SERVER"
    echo "   WORKLOAD_IDENTITY_CLIENT_ID: $WORKLOAD_IDENTITY_CLIENT_ID"
    echo "   KEY_VAULT_NAME: $KEY_VAULT_NAME"
    echo "   TENANT_ID: $TENANT_ID"
    exit 1
fi

echo "✅ Terraform outputs retrieved:"
echo "   ACR: $ACR_LOGIN_SERVER"
echo "   Cluster: $CLUSTER_NAME"
echo "   Key Vault: $KEY_VAULT_NAME"
echo "   DB Host: $DB_HOST"

# Create temporary directory for modified K8s files
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "📋 Preparing Kubernetes manifests..."

# Copy K8s overlay files to temp directory
cp -r "$K8S_OVERLAY_DIR"/* "$TEMP_DIR/"

# Replace placeholders in files
sed -i.bak "s/PLACEHOLDER_ACR_LOGIN_SERVER/$ACR_LOGIN_SERVER/g" "$TEMP_DIR"/*.yaml 2>/dev/null || true
sed -i.bak "s/placeholder-client-id/$WORKLOAD_IDENTITY_CLIENT_ID/g" "$TEMP_DIR"/*.yaml
sed -i.bak "s/placeholder-keyvault/$KEY_VAULT_NAME/g" "$TEMP_DIR"/*.yaml
sed -i.bak "s/placeholder-tenant-id/$TENANT_ID/g" "$TEMP_DIR"/*.yaml
sed -i.bak "s/PLACEHOLDER_DB_HOST/$DB_HOST/g" "$TEMP_DIR"/*.yaml
sed -i.bak "s/PLACEHOLDER_DB_PORT/$DB_PORT/g" "$TEMP_DIR"/*.yaml
sed -i.bak "s/PLACEHOLDER_DB_DATABASE/$DB_DATABASE/g" "$TEMP_DIR"/*.yaml

# Remove backup files
rm -f "$TEMP_DIR"/*.bak

# Get kubectl context
echo "🔧 Configuring kubectl context..."
az aks get-credentials --resource-group $(terraform output -raw resource_group_name 2>/dev/null || echo "hotel-system-rg") --name "$CLUSTER_NAME" --overwrite-existing

# Apply Kubernetes manifests
echo "⚙️  Applying Kubernetes manifests..."

# First, apply the namespace and service account (if needed)
kubectl apply -f "$TEMP_DIR/service-account.yaml"

# Apply the secret provider class
kubectl apply -f "$TEMP_DIR/secret-provider-class.yaml"

# Build the application using kustomize from the temp directory
echo "🏗️  Building and applying with Kustomize..."

# Apply using kustomize from the original overlay directory with substituted files
cd "$K8S_OVERLAY_DIR"

# Copy the modified files back to original overlay directory temporarily
cp "$TEMP_DIR"/*.yaml .

# Apply using kustomize from the overlay directory
kubectl apply -k .

# Clean up - restore original files (optional, since these are likely to be kept)
# Note: We could restore from git if needed

echo "⏳ Waiting for deployment to be ready..."
kubectl rollout status deployment/hotel-system --timeout=30s

echo "🔍 Getting service information..."
kubectl get services hotel-system

echo "✅ Azure deployment completed successfully!"
