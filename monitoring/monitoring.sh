#!/bin/bash

# Barbod Backend Monitoring Stack Management Script
# Usage: ./monitoring.sh [command]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

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

# Create the monitoring network if it doesn't exist
create_network() {
    if ! docker network inspect barbod_monitoring >/dev/null 2>&1; then
        log_info "Creating monitoring network..."
        docker network create barbod_monitoring
    fi
}

# Start monitoring stack
start() {
    log_info "Starting LGTM monitoring stack..."
    create_network
    docker compose -f docker-compose-monitoring.yml up -d
    log_info "Monitoring stack started!"
    log_info ""
    log_info "Access points:"
    log_info "  Grafana:     http://localhost:3000 (admin / check GRAFANA_ADMIN_PASSWORD)"
    log_info "  Prometheus:  http://localhost:9090"
    log_info "  Loki:        http://localhost:3100"
    log_info "  Tempo:       http://localhost:3200"
}

# Start monitoring stack with services (dev)
start_dev() {
    log_info "Starting dev environment with monitoring..."
    create_network
    docker compose -f docker-compose-dev.yml -f docker-compose-monitoring.yml up -d
    log_info "Dev environment with monitoring started!"
}

# Start monitoring stack with services (prod)
start_prod() {
    log_info "Starting production environment with monitoring..."
    create_network
    docker compose -f docker-compose.yml -f docker-compose-monitoring.yml up -d
    log_info "Production environment with monitoring started!"
}

# Stop monitoring stack
stop() {
    log_info "Stopping LGTM monitoring stack..."
    docker compose -f docker-compose-monitoring.yml down
    log_info "Monitoring stack stopped!"
}

# Stop everything
stop_all() {
    log_info "Stopping all services..."
    docker compose -f docker-compose.yml -f docker-compose-monitoring.yml down
    docker compose -f docker-compose-dev.yml -f docker-compose-monitoring.yml down 2>/dev/null || true
    log_info "All services stopped!"
}

# Show status
status() {
    log_info "Monitoring stack status:"
    docker compose -f docker-compose-monitoring.yml ps
}

# View logs
logs() {
    local service=${1:-""}
    if [ -z "$service" ]; then
        docker compose -f docker-compose-monitoring.yml logs -f --tail=100
    else
        docker compose -f docker-compose-monitoring.yml logs -f --tail=100 "$service"
    fi
}

# Health check
health() {
    log_info "Checking health of monitoring services..."
    echo ""
    
    # Check Grafana
    if curl --noproxy '*' -s http://localhost:3000/api/health | grep -q "ok"; then
        echo -e "  Grafana:      ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  Grafana:      ${RED}✗ Unhealthy${NC}"
    fi
    
    # Check Prometheus
    if curl --noproxy '*' -s http://localhost:9090/-/healthy | grep -q "Prometheus"; then
        echo -e "  Prometheus:   ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  Prometheus:   ${RED}✗ Unhealthy${NC}"
    fi
    
    # Check Loki
    if curl --noproxy '*' -s http://localhost:3100/ready | grep -q "ready"; then
        echo -e "  Loki:         ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  Loki:         ${RED}✗ Unhealthy${NC}"
    fi
    
    # Check Tempo
    if curl --noproxy '*' -s http://localhost:3200/ready | grep -q "ready"; then
        echo -e "  Tempo:        ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  Tempo:        ${RED}✗ Unhealthy${NC}"
    fi
    
    # Check OTel Collector
    if curl --noproxy '*' -s http://localhost:13133 | grep -q "Server available"; then
        echo -e "  OTel Collect: ${GREEN}✓ Healthy${NC}"
    else
        echo -e "  OTel Collect: ${RED}✗ Unhealthy${NC}"
    fi
    
    echo ""
}

# Restart monitoring stack
restart() {
    log_info "Restarting LGTM monitoring stack..."
    stop
    start
}

# Pull latest images
update() {
    log_info "Pulling latest monitoring images..."
    docker compose -f docker-compose-monitoring.yml pull
    log_info "Images updated. Run './monitoring.sh restart' to apply."
}

# Clean up volumes (WARNING: destroys data)
clean() {
    log_warn "This will delete all monitoring data!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up monitoring data..."
        docker compose -f docker-compose-monitoring.yml down -v
        log_info "Cleanup complete!"
    else
        log_info "Cleanup cancelled."
    fi
}

# Show help
help() {
    echo "Barbod Backend Monitoring Stack Management"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  start       Start only the monitoring stack"
    echo "  start-dev   Start dev services with monitoring"
    echo "  start-prod  Start production services with monitoring"
    echo "  stop        Stop the monitoring stack"
    echo "  stop-all    Stop all services including monitoring"
    echo "  restart     Restart the monitoring stack"
    echo "  status      Show status of monitoring services"
    echo "  health      Check health of all monitoring services"
    echo "  logs [svc]  View logs (optional: specific service)"
    echo "  update      Pull latest monitoring images"
    echo "  clean       Remove all monitoring data (WARNING!)"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start-dev           # Start everything for development"
    echo "  $0 logs grafana        # View Grafana logs"
    echo "  $0 health              # Check all services health"
}

# Main
case "${1:-help}" in
    start)
        start
        ;;
    start-dev)
        start_dev
        ;;
    start-prod)
        start_prod
        ;;
    stop)
        stop
        ;;
    stop-all)
        stop_all
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    health)
        health
        ;;
    logs)
        logs "$2"
        ;;
    update)
        update
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        help
        ;;
    *)
        log_error "Unknown command: $1"
        help
        exit 1
        ;;
esac
