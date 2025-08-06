# Wili Deployment Guide

This guide explains how to deploy Wili to Yandex Cloud using Kubernetes and GitHub Actions.

## 🏗️ Architecture

- **Frontend**: SvelteKit application (wili.me)
- **Backend**: Two Go microservices
  - User Service: PostgreSQL database
  - Wishlist Service: MongoDB database
- **Infrastructure**: Yandex Managed Kubernetes
- **Domain**: wili.me (with SSL via Let's Encrypt)

## 🔧 Prerequisites

1. **Yandex Cloud Account** with billing enabled
2. **Domain** (wili.me) pointed to your Yandex Cloud load balancer
3. **GitHub Repository** (https://github.com/TheSeems/Wili)
4. **Container Registry** in Yandex Cloud

## 🛠️ Local Development Setup

For local development, you need to set up environment variables:

```bash
# In frontend directory, copy the example environment file
cp .env.example .env

# Edit .env to match your local setup
PUBLIC_API_BASE_URL=http://localhost:8080
```

## 📋 Required GitHub Secrets

You need to configure the following secrets in your GitHub repository settings:

### Yandex Cloud Authentication
- **`YC_SERVICE_ACCOUNT_KEY`**: Service account key in JSON format
- **`YC_CLOUD_ID`**: Your Yandex Cloud ID
- **`YC_FOLDER_ID`**: Your Yandex Cloud folder ID

### Container Registry
- **`YC_REGISTRY_ID`**: Your Yandex Container Registry ID

### Kubernetes Cluster
- **`YC_CLUSTER_NAME`**: Your Kubernetes cluster name
- **`YC_CLUSTER_ID`**: Your Kubernetes cluster ID

## 🚀 Deployment Steps

### 1. Setup Yandex Cloud Resources

```bash
# Install Yandex Cloud CLI
curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash

# Login and configure
yc init

# Create a service account for GitHub Actions
yc iam service-account create --name github-actions

# Assign necessary roles
yc resource-manager folder add-access-binding <FOLDER_ID> \
  --role container-registry.images.puller \
  --subject serviceAccount:<SERVICE_ACCOUNT_ID>

yc resource-manager folder add-access-binding <FOLDER_ID> \
  --role container-registry.images.pusher \
  --subject serviceAccount:<SERVICE_ACCOUNT_ID>

yc resource-manager folder add-access-binding <FOLDER_ID> \
  --role k8s.cluster-api.cluster-admin \
  --subject serviceAccount:<SERVICE_ACCOUNT_ID>

# Create service account key
yc iam key create --service-account-name github-actions --output key.json
```

### 2. Create Container Registry

```bash
yc container registry create --name wili
```

### 3. Create Kubernetes Cluster

```bash
# Create cluster (adjust parameters as needed)
yc managed-kubernetes cluster create \
  --name wili-cluster \
  --network-name default \
  --zone ru-central1-a \
  --subnet-name default-ru-central1-a \
  --public-ip \
  --release-channel stable \
  --version 1.28

# Create node group
yc managed-kubernetes node-group create \
  --name wili-nodes \
  --cluster-name wili-cluster \
  --location zone=ru-central1-a \
  --public-ip \
  --cores 2 \
  --memory 4GB \
  --core-fraction 100 \
  --disk-type network-ssd \
  --disk-size 64GB \
  --fixed-size 2
```

### 4. Setup Domain and SSL

1. Point your domain `wili.me` to the cluster's load balancer IP
2. Install cert-manager in your cluster:

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

3. Install NGINX Ingress Controller:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
```

### 5. Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions, and add:

| Secret Name | Description | How to Get |
|-------------|-------------|------------|
| `YC_SERVICE_ACCOUNT_KEY` | Service account JSON key | Content of `key.json` from step 1 |
| `YC_CLOUD_ID` | Your cloud ID | `yc config list` |
| `YC_FOLDER_ID` | Your folder ID | `yc config list` |
| `YC_REGISTRY_ID` | Container registry ID | `yc container registry list` |
| `YC_CLUSTER_NAME` | Kubernetes cluster name | `wili-cluster` |
| `YC_CLUSTER_ID` | Kubernetes cluster ID | `yc managed-kubernetes cluster list` |

### 6. Deploy

Once secrets are configured, push to the `main` branch to trigger deployment:

```bash
git push origin main
```

## 📊 Manual Deployment

For manual deployment, use the provided script:

```bash
# Set environment variables
export YC_CLOUD_ID="your-cloud-id"
export YC_FOLDER_ID="your-folder-id"
export YC_REGISTRY_ID="your-registry-id"
export YC_CLUSTER_ID="your-cluster-id"

# Run deployment
./scripts/deploy.sh
```

## 🔍 Monitoring and Troubleshooting

### Check Deployment Status
```bash
kubectl get pods -n wili
kubectl get services -n wili
kubectl get ingress -n wili
```

### View Logs
```bash
# Frontend logs
kubectl logs -l app=frontend -n wili

# User service logs
kubectl logs -l app=user-service -n wili

# Wishlist service logs
kubectl logs -l app=wishlist-service -n wili
```

### Scale Services
```bash
# Scale frontend
kubectl scale deployment frontend --replicas=3 -n wili

# Scale user service
kubectl scale deployment user-service --replicas=3 -n wili
```

## 🔒 Security Considerations

1. **Change default passwords** in `k8s/postgres.yaml` and `k8s/mongodb.yaml`
2. **Update JWT secret** in `k8s/user-service.yaml`
3. **Configure proper RBAC** for service accounts
4. **Enable network policies** for pod-to-pod communication
5. **Regular security updates** for base images

## 🌐 URLs

After successful deployment:
- **Frontend**: https://wili.me
- **API**: https://api.wili.me
- **User Service**: https://api.wili.me/api/users
- **Wishlist Service**: https://api.wili.me/api/wishlists

## 📞 Support

For issues with:
- **Yandex Cloud**: Check [Yandex Cloud Documentation](https://yandex.cloud/ru/docs/)
- **Kubernetes**: Check cluster logs and events
- **Application**: Check service logs and GitHub Actions output