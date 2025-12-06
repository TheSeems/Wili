#!/bin/bash

# Template Kubernetes manifests with GitHub secrets
# Usage: ./scripts/template-k8s.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔧 Templating Kubernetes manifests with GitHub secrets...${NC}"

# Create a temporary directory for templated files
TEMP_DIR="k8s-templated"
mkdir -p $TEMP_DIR

# Copy all k8s files to temp directory
cp k8s/*.yaml $TEMP_DIR/

# Replace placeholders with GitHub secrets
echo -e "${YELLOW}📝 Replacing placeholders in Kubernetes manifests...${NC}"

# Database passwords
sed -i "s/DB_PASSWORD_PLACEHOLDER/$DB_PASSWORD/g" $TEMP_DIR/*.yaml
sed -i "s/MONGO_PASSWORD_PLACEHOLDER/$MONGO_PASSWORD/g" $TEMP_DIR/*.yaml

# JWT secret
sed -i "s/JWT_SECRET_PLACEHOLDER/$JWT_SECRET/g" $TEMP_DIR/*.yaml

# Yandex OAuth credentials
sed -i "s/YANDEX_CLIENT_ID_PLACEHOLDER/$YANDEX_CLIENT_ID/g" $TEMP_DIR/*.yaml
sed -i "s/YANDEX_CLIENT_SECRET_PLACEHOLDER/$YANDEX_CLIENT_SECRET/g" $TEMP_DIR/*.yaml

# Registry ID
sed -i "s/REGISTRY_ID_PLACEHOLDER/$YC_REGISTRY_ID/g" $TEMP_DIR/*.yaml

# Image tags
sed -i "s/IMAGE_TAG_PLACEHOLDER/$IMAGE_TAG/g" $TEMP_DIR/*.yaml

# Telegram bot secrets and config
sed -i "s/TELEGRAM_BOT_TOKEN_PLACEHOLDER/$TELEGRAM_BOT_TOKEN/g" $TEMP_DIR/*.yaml
sed -i "s/WEBHOOK_SECRET_TOKEN_PLACEHOLDER/$WEBHOOK_SECRET_TOKEN/g" $TEMP_DIR/*.yaml

echo -e "${GREEN}✅ Kubernetes manifests templated successfully!${NC}"
echo -e "${YELLOW}📁 Templated files are in: $TEMP_DIR/${NC}"

# Apply templated manifests
echo -e "${YELLOW}🚀 Applying templated manifests to Kubernetes...${NC}"

# Apply namespace first
kubectl apply -f $TEMP_DIR/namespace.yaml

# Apply Yandex auth config
kubectl apply -f $TEMP_DIR/yandex-auth-config.yaml

# Apply databases
kubectl apply -f $TEMP_DIR/postgres.yaml
kubectl apply -f $TEMP_DIR/mongodb.yaml

# Wait for databases to be ready
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n wili
kubectl wait --for=condition=available --timeout=300s deployment/mongodb -n wili

# Apply services
kubectl apply -f $TEMP_DIR/user-service.yaml
kubectl apply -f $TEMP_DIR/wishlist-service.yaml
kubectl apply -f $TEMP_DIR/frontend.yaml
kubectl apply -f $TEMP_DIR/telegram-bot.yaml

# Apply ingress
kubectl apply -f $TEMP_DIR/ingress.yaml

# Wait for deployments to be ready
kubectl wait --for=condition=available --timeout=300s deployment/user-service -n wili
kubectl wait --for=condition=available --timeout=300s deployment/wishlist-service -n wili
kubectl wait --for=condition=available --timeout=300s deployment/frontend -n wili
kubectl wait --for=condition=available --timeout=300s deployment/telegram-bot -n wili

echo -e "${GREEN}✅ All manifests applied successfully!${NC}"

# Clean up temp directory
rm -rf $TEMP_DIR

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
