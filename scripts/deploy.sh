#!/bin/bash

# Wili Deployment Script for Yandex Cloud Kubernetes
# This script helps with manual deployment and cluster setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Wili Deployment Script${NC}"
echo "=================================="

# Check if required tools are installed
check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"
    
    if ! command -v yc &> /dev/null; then
        echo -e "${RED}❌ Yandex Cloud CLI is not installed${NC}"
        echo "Install it with: curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}❌ kubectl is not installed${NC}"
        echo "Install it from: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed${NC}"
        echo "Install it from: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All dependencies are installed${NC}"
}

# Setup Yandex Cloud configuration
setup_yc_config() {
    echo -e "${YELLOW}Setting up Yandex Cloud configuration...${NC}"
    
    if [ -z "$YC_CLOUD_ID" ] || [ -z "$YC_FOLDER_ID" ]; then
        echo -e "${RED}❌ Please set YC_CLOUD_ID and YC_FOLDER_ID environment variables${NC}"
        exit 1
    fi
    
    yc config set cloud-id $YC_CLOUD_ID
    yc config set folder-id $YC_FOLDER_ID
    
    echo -e "${GREEN}✅ Yandex Cloud configured${NC}"
}

# Configure Docker for Yandex Container Registry
setup_docker() {
    echo -e "${YELLOW}Configuring Docker for Yandex Container Registry...${NC}"
    yc container registry configure-docker
    echo -e "${GREEN}✅ Docker configured${NC}"
}

# Build and push images
build_and_push() {
    echo -e "${YELLOW}Building and pushing Docker images...${NC}"
    
    if [ -z "$YC_REGISTRY_ID" ]; then
        echo -e "${RED}❌ Please set YC_REGISTRY_ID environment variable${NC}"
        exit 1
    fi
    
    REGISTRY="cr.yandex/$YC_REGISTRY_ID"
    TAG=${1:-latest}
    
    # Build User Service
    echo "Building user service..."
    cd backend/services/user
    docker build -t $REGISTRY/user-service:$TAG .
    docker push $REGISTRY/user-service:$TAG
    cd ../../..
    
    # Build Wishlist Service
    echo "Building wishlist service..."
    cd backend/services/wishlist
    docker build -t $REGISTRY/wishlist-service:$TAG .
    docker push $REGISTRY/wishlist-service:$TAG
    cd ../../..
    
    # Build Frontend
    echo "Building frontend..."
    cd frontend
    docker build -t $REGISTRY/frontend:$TAG .
    docker push $REGISTRY/frontend:$TAG
    cd ..
    
    echo -e "${GREEN}✅ All images built and pushed${NC}"
}

# Deploy to Kubernetes
deploy_k8s() {
    echo -e "${YELLOW}Deploying to Kubernetes...${NC}"
    
    if [ -z "$YC_CLUSTER_ID" ]; then
        echo -e "${RED}❌ Please set YC_CLUSTER_ID environment variable${NC}"
        exit 1
    fi
    
    # Get cluster credentials
    yc managed-kubernetes cluster get-credentials $YC_CLUSTER_ID --external
    
    # Apply manifests
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/postgres.yaml
    kubectl apply -f k8s/mongodb.yaml
    
    # Wait for databases
    echo "Waiting for databases to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n wili
    kubectl wait --for=condition=available --timeout=300s deployment/mongodb -n wili
    
    # Deploy services
    kubectl apply -f k8s/user-service.yaml
    kubectl apply -f k8s/wishlist-service.yaml
    kubectl apply -f k8s/frontend.yaml
    kubectl apply -f k8s/ingress.yaml
    
    # Wait for services
    echo "Waiting for services to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/user-service -n wili
    kubectl wait --for=condition=available --timeout=300s deployment/wishlist-service -n wili
    kubectl wait --for=condition=available --timeout=300s deployment/frontend -n wili
    
    echo -e "${GREEN}✅ Deployment completed${NC}"
}

# Show status
show_status() {
    echo -e "${YELLOW}Deployment Status:${NC}"
    kubectl get pods -n wili
    echo ""
    kubectl get services -n wili
    echo ""
    kubectl get ingress -n wili
}

# Main execution
case "${1:-all}" in
    "deps")
        check_dependencies
        ;;
    "config")
        check_dependencies
        setup_yc_config
        setup_docker
        ;;
    "build")
        check_dependencies
        setup_yc_config
        setup_docker
        build_and_push $2
        ;;
    "deploy")
        check_dependencies
        setup_yc_config
        deploy_k8s
        ;;
    "status")
        show_status
        ;;
    "all")
        check_dependencies
        setup_yc_config
        setup_docker
        build_and_push latest
        deploy_k8s
        show_status
        ;;
    *)
        echo "Usage: $0 {deps|config|build|deploy|status|all}"
        echo ""
        echo "Commands:"
        echo "  deps    - Check dependencies"
        echo "  config  - Setup Yandex Cloud configuration"
        echo "  build   - Build and push Docker images"
        echo "  deploy  - Deploy to Kubernetes"
        echo "  status  - Show deployment status"
        echo "  all     - Run complete deployment (default)"
        exit 1
        ;;
esac

echo -e "${GREEN}🎉 Script completed successfully!${NC}"