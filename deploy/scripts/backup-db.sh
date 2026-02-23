#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  backup-db.sh — PostgreSQL Database Backup                  ║
# ║                                                              ║
# ║  Usage:                                                      ║
# ║    ./scripts/backup-db.sh                  # Default backup  ║
# ║    ./scripts/backup-db.sh /custom/path     # Custom path     ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

# ── Configuration ────────────────────────────────────────────
CONTAINER_NAME="atstex-postgres"
DB_USER="${POSTGRES_USER:-atstex}"
DB_NAME="${POSTGRES_DB:-atstex}"
BACKUP_DIR="${1:-/opt/atstex-lab/backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/atstex_${TIMESTAMP}.sql.gz"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# ── Colors ───────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}📦 Starting PostgreSQL backup...${NC}"

# ── Pre-flight ───────────────────────────────────────────────
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}❌ Container '${CONTAINER_NAME}' is not running!${NC}"
    exit 1
fi

mkdir -p "${BACKUP_DIR}"

# ── Backup ───────────────────────────────────────────────────
echo "   Database: ${DB_NAME}"
echo "   Output:   ${BACKUP_FILE}"

docker exec "${CONTAINER_NAME}" \
    pg_dump -U "${DB_USER}" -d "${DB_NAME}" \
    --verbose --clean --if-exists --create \
    2>/dev/null | gzip > "${BACKUP_FILE}"

# ── Verify ───────────────────────────────────────────────────
BACKUP_SIZE=$(ls -lh "${BACKUP_FILE}" | awk '{print $5}')
echo -e "${GREEN}✅ Backup complete: ${BACKUP_FILE} (${BACKUP_SIZE})${NC}"

# ── Cleanup Old Backups ─────────────────────────────────────
DELETED=$(find "${BACKUP_DIR}" -name "atstex_*.sql.gz" -mtime +${RETENTION_DAYS} -delete -print | wc -l)
if [[ ${DELETED} -gt 0 ]]; then
    echo -e "${YELLOW}🧹 Removed ${DELETED} backup(s) older than ${RETENTION_DAYS} days${NC}"
fi

# ── List Recent Backups ──────────────────────────────────────
echo ""
echo "Recent backups:"
ls -lhrt "${BACKUP_DIR}"/atstex_*.sql.gz 2>/dev/null | tail -5
