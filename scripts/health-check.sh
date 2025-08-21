#!/bin/bash

# Ganje Health Check Script
set -e

# Configuration
NAMESPACE=${NAMESPACE:-default}
RELEASE_NAME=${RELEASE_NAME:-ganje}
TIMEOUT=${TIMEOUT:-300}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_pods() {
    log_info "Checking pod status..."
    
    PODS=$(kubectl get pods -l app.kubernetes.io/name=ganje -n ${NAMESPACE} --no-headers)
    if [ -z "$PODS" ]; then
        log_error "No Ganje pods found"
        return 1
    fi
    
    echo "$PODS"
    
    # Check if all pods are ready
    NOT_READY=$(kubectl get pods -l app.kubernetes.io/name=ganje -n ${NAMESPACE} --no-headers | grep -v "Running\|Completed" || true)
    if [ -n "$NOT_READY" ]; then
        log_warn "Some pods are not ready:"
        echo "$NOT_READY"
        return 1
    fi
    
    log_info "All pods are running"
    return 0
}

check_services() {
    log_info "Checking service status..."
    
    SERVICE=$(kubectl get svc ${RELEASE_NAME} -n ${NAMESPACE} --no-headers 2>/dev/null || echo "")
    if [ -z "$SERVICE" ]; then
        log_error "Service ${RELEASE_NAME} not found"
        return 1
    fi
    
    echo "$SERVICE"
    log_info "Service is available"
    return 0
}

check_ingress() {
    log_info "Checking ingress status..."
    
    INGRESS=$(kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} --no-headers 2>/dev/null || echo "")
    if [ -z "$INGRESS" ]; then
        log_warn "No ingress found"
        return 0
    fi
    
    echo "$INGRESS"
    
    # Check if ingress has an address
    ADDRESS=$(kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ -z "$ADDRESS" ]; then
        ADDRESS=$(kubectl get ingress ${RELEASE_NAME} -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    fi
    
    if [ -n "$ADDRESS" ]; then
        log_info "Ingress address: $ADDRESS"
    else
        log_warn "Ingress has no external address yet"
    fi
    
    return 0
}

check_pvc() {
    log_info "Checking persistent volume claims..."
    
    PVC=$(kubectl get pvc ${RELEASE_NAME}-storage -n ${NAMESPACE} --no-headers 2>/dev/null || echo "")
    if [ -z "$PVC" ]; then
        log_warn "No PVC found"
        return 0
    fi
    
    echo "$PVC"
    
    # Check PVC status
    STATUS=$(kubectl get pvc ${RELEASE_NAME}-storage -n ${NAMESPACE} -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
    if [ "$STATUS" != "Bound" ]; then
        log_warn "PVC is not bound (status: $STATUS)"
        return 1
    fi
    
    log_info "PVC is bound"
    return 0
}

test_health_endpoint() {
    log_info "Testing health endpoint..."
    
    # Port forward to test health endpoint
    kubectl port-forward svc/${RELEASE_NAME} 8080:8080 -n ${NAMESPACE} &
    PF_PID=$!
    
    # Wait for port forward to be ready
    sleep 5
    
    # Test health endpoint
    if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
        log_info "Health endpoint is responding"
        HEALTH_OK=true
    else
        log_error "Health endpoint is not responding"
        HEALTH_OK=false
    fi
    
    # Clean up port forward
    kill $PF_PID 2>/dev/null || true
    
    if [ "$HEALTH_OK" = true ]; then
        return 0
    else
        return 1
    fi
}

test_repository_endpoints() {
    log_info "Testing repository endpoints..."
    
    # Port forward to test repository endpoints
    kubectl port-forward svc/${RELEASE_NAME} 8080:8080 -n ${NAMESPACE} &
    PF_PID=$!
    
    # Wait for port forward to be ready
    sleep 5
    
    ENDPOINTS=("maven-local" "npm-local" "docker-local" "pypi-local")
    ALL_OK=true
    
    for endpoint in "${ENDPOINTS[@]}"; do
        if curl -f -s "http://localhost:8080/${endpoint}/" > /dev/null 2>&1; then
            log_info "Repository endpoint /${endpoint}/ is responding"
        else
            log_warn "Repository endpoint /${endpoint}/ is not responding"
            ALL_OK=false
        fi
    done
    
    # Clean up port forward
    kill $PF_PID 2>/dev/null || true
    
    if [ "$ALL_OK" = true ]; then
        log_info "All repository endpoints are responding"
        return 0
    else
        log_warn "Some repository endpoints are not responding"
        return 1
    fi
}

show_logs() {
    log_info "Recent application logs:"
    kubectl logs -l app.kubernetes.io/name=ganje -n ${NAMESPACE} --tail=20 --prefix=true
}

show_events() {
    log_info "Recent events:"
    kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' --field-selector involvedObject.kind=Pod | tail -10
}

comprehensive_check() {
    log_info "Running comprehensive health check..."
    
    CHECKS_PASSED=0
    TOTAL_CHECKS=5
    
    # Check pods
    if check_pods; then
        ((CHECKS_PASSED++))
    fi
    
    # Check services
    if check_services; then
        ((CHECKS_PASSED++))
    fi
    
    # Check ingress
    if check_ingress; then
        ((CHECKS_PASSED++))
    fi
    
    # Check PVC
    if check_pvc; then
        ((CHECKS_PASSED++))
    fi
    
    # Test health endpoint
    if test_health_endpoint; then
        ((CHECKS_PASSED++))
    fi
    
    echo ""
    log_info "Health check summary: $CHECKS_PASSED/$TOTAL_CHECKS checks passed"
    
    if [ $CHECKS_PASSED -eq $TOTAL_CHECKS ]; then
        log_info "All health checks passed! ✅"
        return 0
    else
        log_warn "Some health checks failed! ⚠️"
        return 1
    fi
}

# Main script
case "${1:-check}" in
    "pods")
        check_pods
        ;;
    "services")
        check_services
        ;;
    "ingress")
        check_ingress
        ;;
    "pvc")
        check_pvc
        ;;
    "health")
        test_health_endpoint
        ;;
    "endpoints")
        test_repository_endpoints
        ;;
    "logs")
        show_logs
        ;;
    "events")
        show_events
        ;;
    "check"|"all")
        comprehensive_check
        ;;
    *)
        echo "Usage: $0 {pods|services|ingress|pvc|health|endpoints|logs|events|check}"
        echo ""
        echo "Commands:"
        echo "  pods      - Check pod status"
        echo "  services  - Check service status"
        echo "  ingress   - Check ingress status"
        echo "  pvc       - Check persistent volume claims"
        echo "  health    - Test health endpoint"
        echo "  endpoints - Test repository endpoints"
        echo "  logs      - Show recent logs"
        echo "  events    - Show recent events"
        echo "  check     - Run comprehensive health check"
        echo ""
        echo "Environment Variables:"
        echo "  NAMESPACE     - Kubernetes namespace (default: default)"
        echo "  RELEASE_NAME  - Helm release name (default: ganje)"
        echo "  TIMEOUT       - Timeout in seconds (default: 300)"
        exit 1
        ;;
esac
