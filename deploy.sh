#!/bin/bash
# Deploy the e-learning platform to the server over ONE ssh connection
# (password asked only once).
#
# Usage (Git Bash / WSL / Linux / macOS, from the project folder):
#   bash deploy.sh            # normal deploy (uses Docker build cache)
#   bash deploy.sh --clean    # rebuild images from scratch
#
# Reads from .env:
#   DEPLOY_HOST        server IP/hostname       (fallback: DB_HOST)
#   DEPLOY_USER        ssh user                 (fallback: DB_USER)
#   DEPLOY_REMOTE_DIR  folder on the server     (default: /home/<user>/e-learning)
#   APP_DOMAIN         your domain, only used in the final message
# The local .env is copied to the server (Windows line endings removed).

set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f ".env" ]; then
    echo "ERROR: .env not found. Run: cp .env.example .env  and fill it in."
    exit 1
fi

# Read one value from .env without executing it: strips \r and surrounding quotes.
env_get() {
    { grep -E "^[[:space:]]*$1[[:space:]]*=" .env || true; } | tail -n 1 | cut -d= -f2- \
        | tr -d '\r' | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'$/\1/"
}

DEPLOY_HOST="$(env_get DEPLOY_HOST)"; DEPLOY_HOST="${DEPLOY_HOST:-$(env_get DB_HOST)}"
DEPLOY_USER="$(env_get DEPLOY_USER)"; DEPLOY_USER="${DEPLOY_USER:-$(env_get DB_USER)}"
REMOTE_DIR="$(env_get DEPLOY_REMOTE_DIR)"; REMOTE_DIR="${REMOTE_DIR:-/home/${DEPLOY_USER}/e-learning}"
APP_DOMAIN="$(env_get APP_DOMAIN)"; APP_DOMAIN="${APP_DOMAIN:-$DEPLOY_HOST}"
BUILD_FLAGS=""
[ "${1:-}" = "--clean" ] && BUILD_FLAGS="--no-cache"

# ---- sanity checks before touching the server
fail=0
[ -z "$DEPLOY_HOST" ] && { echo "ERROR: set DEPLOY_HOST in .env"; fail=1; }
[ -z "$DEPLOY_USER" ] && { echo "ERROR: set DEPLOY_USER in .env"; fail=1; }
jwt="$(env_get JWT_SECRET)"
if [ ${#jwt} -lt 32 ] || [[ "$jwt" == *change-this* ]]; then
    echo "ERROR: JWT_SECRET must be at least 32 random characters (got ${#jwt}). Create one with: openssl rand -hex 32"
    fail=1
fi
for v in DB_USER DB_PASSWORD DB_NAME; do
    [ -z "$(env_get $v)" ] && { echo "ERROR: $v is empty in .env"; fail=1; }
done
[ $fail -eq 1 ] && exit 1

SERVER="${DEPLOY_USER}@${DEPLOY_HOST}"
ENV_CONTENT="$(tr -d '\r' < .env)"

echo "=== Deploying E-Learning ==="
echo "Server: $SERVER"
echo "Folder: $REMOTE_DIR"
echo "(enter the ssh password once if asked)"
echo ""

tar czf - \
    --exclude=".git" --exclude="node_modules" --exclude=".env" --exclude="dist" \
    --exclude="go.dev.mod" --exclude="go.dev.sum" \
    api web proxy docker-compose.yaml .env.example README.md docs | \
ssh "$SERVER" "
set -e
mkdir -p '$REMOTE_DIR'
cd '$REMOTE_DIR'

echo '[1/4] Extracting files...'
cat > /tmp/elearn_deploy.tar.gz
tar xzf /tmp/elearn_deploy.tar.gz
rm -f /tmp/elearn_deploy.tar.gz

echo '[2/4] Writing .env...'
cat > .env <<'ENVEOF'
$ENV_CONTENT
ENVEOF
chmod 600 .env

echo '[3/4] Building and starting containers...'
docker compose build $BUILD_FLAGS
docker compose up -d --remove-orphans

echo '[4/4] Waiting for the API (max 90s)...'
ok=0
for i in \$(seq 1 30); do
    code=\$(curl -s -o /dev/null -w '%{http_code}' http://localhost/health || true)
    if [ \"\$code\" = '200' ]; then ok=1; break; fi
    sleep 3
done
docker compose ps
echo ''
if [ \$ok -eq 1 ]; then
    echo '✅ Deploy OK'
    echo '   Website:  https://$APP_DOMAIN  (or http://$DEPLOY_HOST)'
    echo '   Device:   https://$APP_DOMAIN/api'
else
    echo '❌ API is not healthy. Last API logs:'
    docker compose logs --tail 40 api
    exit 1
fi
"
