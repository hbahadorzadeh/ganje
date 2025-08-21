#!/bin/bash

# Ganje Deployment Script
set -e

# Configuration
NAMESPACE=${NAMESPACE:-default}
RELEASE_NAME=${RELEASE_NAME:-ganje}
CHART_PATH=${CHART_PATH:-./helm/ganje}
VALUES_FILE=${VALUES_FILE:-values.yaml}
IMAGE_TAG=${IMAGE_TAG:-latest}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed"
        exit 1
    fi
    
    # Check if docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "docker is not installed"
        exit 1
    fi
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

build_image() {
    log_info "Building Docker image..."
    docker build -t ganje:${IMAGE_TAG} .
    log_info "Docker image built successfully"
}

deploy_dependencies() {
    log_info "Deploying dependencies..."
    
    # Add Bitnami repo if not exists
    if ! helm repo list | grep -q bitnami; then
        helm repo add bitnami https://charts.bitnami.com/bitnami
    fi
    
    helm repo update
    
    # Check if PostgreSQL is already deployed
    if ! helm list -n ${NAMESPACE} | grep -q postgresql; then
        log_info "Deploying PostgreSQL..."
        helm install postgresql bitnami/postgresql \
            --namespace ${NAMESPACE} \
            --create-namespace \
            --set auth.postgresPassword=ganje-password \
            --set auth.database=ganje \
            --set auth.username=ganje \
            --set auth.password=ganje-password \
            --set primary.persistence.size=20Gi
        
        log_info "Waiting for PostgreSQL to be ready..."
        kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=postgresql -n ${NAMESPACE} --timeout=300s
    else
        log_info "PostgreSQL already deployed"
    fi
}

deploy_ganje() {
    log_info "Deploying Ganje..."
    
    # Create namespace if it doesn't exist
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Check if release exists
    if helm list -n ${NAMESPACE} | grep -q ${RELEASE_NAME}; then
        log_info "Upgrading existing release..."
        helm upgrade ${RELEASE_NAME} ${CHART_PATH} \
            --namespace ${NAMESPACE} \
            --values ${CHART_PATH}/${VALUES_FILE} \
            --set image.tag=${IMAGE_TAG} \
            --wait
    else
        log_info "Installing new release..."
        helm install ${RELEASE_NAME} ${CHART_PATH} \
            --namespace ${NAMESPACE} \
            --values ${CHART_PATH}/${VALUES_FILE} \
            --set image.tag=${IMAGE_TAG} \
            --wait
    fi
    
    log_info "Ganje deployed successfully"
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check if pods are running
    kubectl get pods -l app.kubernetes.io/name=ganje -n ${NAMESPACE}
    
    # Wait for pods to be ready
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ganje -n ${NAMESPACE} --timeout=300s
    
    # Check service
    kubectl get svc ${RELEASE_NAME} -n ${NAMESPACE}
    
    # Check ingress if enabled
    if kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} &> /dev/null; then
        kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE}
    fi
    
    log_info "Deployment verification completed"
}

show_access_info() {
    log_info "Access Information:"
    
    # Get ingress info
    if kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} &> /dev/null; then
        INGRESS_HOST=$(kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.rules[0].host}')
        echo "  Ingress URL: http://${INGRESS_HOST}"
    fi
    
    # Port forward command
    echo "  Port Forward: kubectl port-forward svc/${RELEASE_NAME} 8080:8080 -n ${NAMESPACE}"
    
    # Repository endpoints
    echo "  Repository Endpoints:"
    echo "    Maven: /maven-local/"
    echo "    NPM: /npm-local/"
    echo "    Docker: /docker-local/"
    echo "    PyPI: /pypi-local/"
}

cleanup() {
    log_info "Cleaning up deployment..."
    
    # Uninstall Ganje
    if helm list -n ${NAMESPACE} | grep -q ${RELEASE_NAME}; then
        helm uninstall ${RELEASE_NAME} -n ${NAMESPACE}
    fi
    
    # Optionally remove PostgreSQL
    read -p "Remove PostgreSQL? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if helm list -n ${NAMESPACE} | grep -q postgresql; then
            helm uninstall postgresql -n ${NAMESPACE}
        fi
    fi
    
    log_info "Cleanup completed"
}

# Main script
case "${1:-deploy}" in
    "build")
        check_prerequisites
        build_image
        ;;
    "deploy")
        check_prerequisites
        build_image
        deploy_dependencies
        deploy_ganje
        verify_deployment
        show_access_info
        ;;
    "upgrade")
        check_prerequisites
        build_image
        deploy_ganje
        verify_deployment
        ;;
    "verify")
        verify_deployment
        show_access_info
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        echo "Usage: $0 {build|deploy|upgrade|verify|cleanup}"
        echo ""
        echo "Commands:"
        echo "  build    - Build Docker image only"
        echo "  deploy   - Full deployment (build + dependencies + deploy)"
        echo "  upgrade  - Upgrade existing deployment"
        echo "  verify   - Verify deployment status"
        echo "  cleanup  - Remove deployment"
        echo ""
        echo "Environment Variables:"
        echo "  NAMESPACE     - Kubernetes namespace (default: default)"
        echo "  RELEASE_NAME  - Helm release name (default: ganje)"
        echo "  CHART_PATH    - Path to Helm chart (default: ./helm/ganje)"
        echo "  VALUES_FILE   - Values file name (default: values.yaml)"
        echo "  IMAGE_TAG     - Docker image tag (default: latest)"
        exit 1
        ;;
esac
