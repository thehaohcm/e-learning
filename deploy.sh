#!/bin/bash
# Deploy e-learning platform to remote server
# Reads credentials from .env file
# Usage: bash deploy.sh

set -e

# Load .env file
if [ ! -f ".env" ]; then
    echo "ERROR: .env file not found!"
    echo "  cp .env.example .env"
    echo "  Then edit .env with your real credentials"
    exit 1
fi

# Source .env (only exports variables, safe for bash)
export $(grep -v '^#' .env | grep -v '^$' | xargs)

# Use DEPLOY_* vars for SSH connection (or fallback to DB creds)
DEPLOY_HOST="${DEPLOY_HOST:-$DB_HOST}"
DEPLOY_USER="${DEPLOY_USER:-$DB_USER}"
SERVER="${DEPLOY_USER}@${DEPLOY_HOST}"
REMOTE_DIR="${DEPLOY_REMOTE_DIR:-/home/${DEPLOY_USER}/e-learning}"
API_PORT="${PORT:-5173}"
FRONTEND_PORT="${FRONTEND_PORT:-8085}"

echo "=== Deploying E-Learning Platform ==="
echo "Server: $SERVER"
echo ""

# 1. Create remote directory
echo "[1/4] Creating remote directory..."
ssh "$SERVER" "mkdir -p $REMOTE_DIR/uploads"

# 2. Copy source files
echo "[2/4] Copying source files..."
scp -r api web docker-compose.yaml .env "$SERVER:$REMOTE_DIR/"

# 3. Build and start with Docker Compose
echo "[3/4] Building and starting Docker services..."
ssh "$SERVER" "cd $REMOTE_DIR && docker compose down 2>/dev/null; docker compose build --no-cache && docker compose up -d"

# 4. Verify
echo "[4/4] Verifying..."
sleep 8
echo ""
ssh "$SERVER" "cd $REMOTE_DIR && docker compose ps"
echo ""

# Health check
HTTP_CODE=$(ssh "$SERVER" "curl -s -o /dev/null -w '%{http_code}' http://localhost:${API_PORT}/health 2>/dev/null || echo '000'")
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Backend is healthy (HTTP $HTTP_CODE)"
else
    echo "⏳ Backend starting... (HTTP $HTTP_CODE) - check logs:"
    echo "   ssh $SERVER 'cd $REMOTE_DIR && docker compose logs api'"
fi

echo ""
echo "============================================"
echo "  DEPLOYMENT COMPLETE"
echo "============================================"
echo ""
echo "  Frontend:  http://${DEPLOY_HOST}:${FRONTEND_PORT}"
echo "  Backend:   http://${DEPLOY_HOST}:${API_PORT}"
echo "  Health:    http://${DEPLOY_HOST}:${API_PORT}/health"
echo ""
echo "  View logs:"
echo "    ssh $SERVER 'cd $REMOTE_DIR && docker compose logs -f'"
echo ""
echo "  Stop server:"
echo "    ssh $SERVER 'cd $REMOTE_DIR && docker compose down'"
echo ""