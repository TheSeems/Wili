# GitHub Secrets Configuration

This document describes the required GitHub secrets for the Wili deployment pipeline.

## Required Secrets

### Database Secrets
- `DB_PASSWORD`: Base64 encoded PostgreSQL password
- `MONGO_PASSWORD`: Base64 encoded MongoDB password

### Authentication Secrets
- `JWT_SECRET`: Base64 encoded JWT signing key for user authentication

### Yandex OAuth Secrets
- `YANDEX_CLIENT_ID`: Yandex OAuth application client ID
- `YANDEX_CLIENT_SECRET`: Yandex OAuth application client secret

### Yandex Cloud Secrets
- `YC_REGISTRY_ID`: Yandex Container Registry ID
- `YC_CLOUD_ID`: Yandex Cloud ID
- `YC_FOLDER_ID`: Yandex Cloud folder ID
- `YC_CLUSTER_ID`: Yandex Managed Kubernetes cluster ID
- `YC_CLUSTER_NAME`: Yandex Managed Kubernetes cluster name
- `YC_SERVICE_ACCOUNT_KEY`: JSON service account key for Yandex Cloud

## Setting Up Secrets

1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add each secret with the appropriate value

## Base64 Encoding

For database passwords and JWT secrets, you need to base64 encode the values:

```bash
# For database passwords
echo -n "your_password" | base64

# For JWT secret
echo -n "your_jwt_secret" | base64
```

## Example Values

```bash
# Database passwords (example)
DB_PASSWORD=$(echo -n "postgres" | base64)
MONGO_PASSWORD=$(echo -n "password" | base64)

# JWT secret (example - use a strong secret in production)
JWT_SECRET=$(echo -n "my_super_secret_jwt_key" | base64)

# Yandex OAuth (from your Yandex OAuth application)
YANDEX_CLIENT_ID="23a421100b584048b2265ef34ab4b933"
YANDEX_CLIENT_SECRET="9f776a0c182f478fb27496631757a959"

# Yandex Cloud (from your Yandex Cloud console)
YC_REGISTRY_ID="crp68lk09pak2nimrv24"
YC_CLOUD_ID="your_cloud_id"
YC_FOLDER_ID="your_folder_id"
YC_CLUSTER_ID="your_cluster_id"
YC_CLUSTER_NAME="your_cluster_name"
YC_SERVICE_ACCOUNT_KEY='{"your":"service_account_json"}'
```

## Security Notes

- Never commit secrets to the repository
- Use strong, unique passwords for databases
- Use a cryptographically secure random string for JWT_SECRET
- Rotate secrets regularly
- Use different secrets for different environments (staging/production)

## Template System

The deployment pipeline uses a templating system that replaces placeholders in Kubernetes manifests with GitHub secrets:

- `DB_PASSWORD_PLACEHOLDER` → `$DB_PASSWORD`
- `MONGO_PASSWORD_PLACEHOLDER` → `$MONGO_PASSWORD`
- `JWT_SECRET_PLACEHOLDER` → `$JWT_SECRET`
- `YANDEX_CLIENT_ID_PLACEHOLDER` → `$YANDEX_CLIENT_ID`
- `YANDEX_CLIENT_SECRET_PLACEHOLDER` → `$YANDEX_CLIENT_SECRET`
- `REGISTRY_ID_PLACEHOLDER` → `$YC_REGISTRY_ID`
- `IMAGE_TAG_PLACEHOLDER` → `$IMAGE_TAG` (GitHub SHA)
