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

# Helper function to base64 encode (works on both Linux and macOS)
base64_encode() {
    if command -v base64 >/dev/null 2>&1; then
        echo -n "$1" | base64 | tr -d '\n'
    else
        echo "Error: base64 command not found" >&2
        exit 1
    fi
}

# Database passwords (base64 encode for Kubernetes secrets)
DB_PASSWORD_B64=$(base64_encode "$DB_PASSWORD")
MONGO_PASSWORD_B64=$(base64_encode "$MONGO_PASSWORD")
sed -i "s|DB_PASSWORD_PLACEHOLDER|$DB_PASSWORD_B64|g" $TEMP_DIR/*.yaml
sed -i "s|MONGO_PASSWORD_PLACEHOLDER|$MONGO_PASSWORD_B64|g" $TEMP_DIR/*.yaml

# JWT secret (base64 encode for Kubernetes secrets)
JWT_SECRET_B64=$(base64_encode "$JWT_SECRET")
sed -i "s|JWT_SECRET_PLACEHOLDER|$JWT_SECRET_B64|g" $TEMP_DIR/*.yaml

# Yandex OAuth credentials (base64 encode for Kubernetes secrets)
YANDEX_CLIENT_ID_B64=$(base64_encode "$YANDEX_CLIENT_ID")
YANDEX_CLIENT_SECRET_B64=$(base64_encode "$YANDEX_CLIENT_SECRET")
sed -i "s|YANDEX_CLIENT_ID_PLACEHOLDER|$YANDEX_CLIENT_ID_B64|g" $TEMP_DIR/*.yaml
sed -i "s|YANDEX_CLIENT_SECRET_PLACEHOLDER|$YANDEX_CLIENT_SECRET_B64|g" $TEMP_DIR/*.yaml

# Registry ID (not a secret, plain text)
sed -i "s|REGISTRY_ID_PLACEHOLDER|$YC_REGISTRY_ID|g" $TEMP_DIR/*.yaml

# Image tags (not a secret, plain text)
sed -i "s|IMAGE_TAG_PLACEHOLDER|$IMAGE_TAG|g" $TEMP_DIR/*.yaml

# Telegram bot secrets (base64 encode for Kubernetes secrets)
TELEGRAM_BOT_TOKEN_B64=$(base64_encode "$TELEGRAM_BOT_TOKEN")
WEBHOOK_SECRET_TOKEN_B64=$(base64_encode "$WEBHOOK_SECRET_TOKEN")
sed -i "s|TELEGRAM_BOT_TOKEN_PLACEHOLDER|$TELEGRAM_BOT_TOKEN_B64|g" $TEMP_DIR/*.yaml
sed -i "s|WEBHOOK_SECRET_TOKEN_PLACEHOLDER|$WEBHOOK_SECRET_TOKEN_B64|g" $TEMP_DIR/*.yaml

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

# Set Telegram webhook after bot is ready
echo -e "${YELLOW}🔗 Setting Telegram webhook...${NC}"
WEBHOOK_URL="https://tg.wili.me/webhook"

# Set webhook with secret token
# Note: TELEGRAM_BOT_TOKEN and WEBHOOK_SECRET_TOKEN are already in plain text from GitHub secrets
WEBHOOK_RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"${WEBHOOK_URL}\",\"secret_token\":\"${WEBHOOK_SECRET_TOKEN}\",\"allowed_updates\":[\"message\",\"inline_query\"]}")

if echo "$WEBHOOK_RESPONSE" | grep -q '"ok":true'; then
  echo -e "${GREEN}✅ Telegram webhook set successfully${NC}"
else
  echo -e "${YELLOW}⚠️  Failed to set Telegram webhook: ${WEBHOOK_RESPONSE}${NC}"
  echo -e "${YELLOW}⚠️  You may need to set it manually:${NC}"
  echo -e "${YELLOW}   curl -X POST \"https://api.telegram.org/bot\${TELEGRAM_BOT_TOKEN}/setWebhook\" -d \"url=${WEBHOOK_URL}&secret_token=\${WEBHOOK_SECRET_TOKEN}\"${NC}"
fi

echo -e "${GREEN}✅ All manifests applied successfully!${NC}"

# Clean up temp directory
rm -rf $TEMP_DIR

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
