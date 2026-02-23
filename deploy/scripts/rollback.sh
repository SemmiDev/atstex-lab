#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  rollback.sh — Rollback to Previous Docker Image            ║
# ║                                                              ║
# ║  Usage:                                                      ║
# ║    ./scripts/rollback.sh                                     ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

# ── Configuration ────────────────────────────────────────────
APP_NAME="${APP_NAME:-atstex-lab}"
APP_DIR="${APP_DIR:-/opt/atstex-lab}"
COMPOSE_FILE="${APP_DIR}/docker-compose-app.yml"

# ── Colors ───────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}╔══════════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║  ATSTEX-LAB — Rollback                       ║${NC}"
echo -e "${YELLOW}╚══════════════════════════════════════════════╝${NC}"
echo ""

# ── Check rollback image exists ─────────────────────────────
ROLLBACK_IMAGE=$(docker images -q "${APP_NAME}:rollback" 2>/dev/null)

if [[ -z "${ROLLBACK_IMAGE}" ]]; then
    echo -e "${RED}❌ No rollback image found (${APP_NAME}:rollback)${NC}"
    echo "   Rollback is only available if a previous deployment was done."
    echo ""
    echo "Available images:"
    docker images "${APP_NAME}" --format "table {{.Tag}}\t{{.CreatedAt}}\t{{.Size}}"
    exit 1
fi

CURRENT_IMAGE=$(docker images -q "${APP_NAME}:latest" 2>/dev/null)

echo "Current image:  ${CURRENT_IMAGE:-none}"
echo "Rollback image: ${ROLLBACK_IMAGE}"
echo ""

read -p "$(echo -e ${YELLOW}"▶ Rollback to previous version? (y/N): "${NC})" confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo -e "${RED}❌ Rollback cancelled.${NC}"
    exit 0
fi

# ── Perform Rollback ────────────────────────────────────────
echo ""
echo -e "${YELLOW}🔄 Rolling back...${NC}"

# Stop current container
docker compose -f "${COMPOSE_FILE}" down 2>/dev/null || true

# Swap images
docker tag "${APP_NAME}:latest" "${APP_NAME}:failed" 2>/dev/null || true
docker tag "${APP_NAME}:rollback" "${APP_NAME}:latest"

# Start with rollback image
docker compose -f "${COMPOSE_FILE}" up -d

# ── Health Check ─────────────────────────────────────────────
echo ""
echo -e "${YELLOW}⏳ Waiting for application to start...${NC}"

RETRIES=12
for i in $(seq 1 $RETRIES); do
    if curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:8080" | grep -qE "^(200|301|302)$"; then
        echo -e "${GREEN}✅ Rollback successful! Application is running.${NC}"
        echo ""
        echo "Rollback summary:"
        docker images "${APP_NAME}" --format "table {{.Tag}}\t{{.CreatedAt}}\t{{.Size}}"
        exit 0
    fi
    echo "   Attempt $i/$RETRIES — waiting..."
    sleep 5
done

echo -e "${RED}❌ Application failed to start after rollback!${NC}"
echo "   Check logs: docker compose -f ${COMPOSE_FILE} logs --tail 50"
exit 1
