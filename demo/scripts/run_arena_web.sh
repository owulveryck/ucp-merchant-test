#!/bin/bash
# Remote Arena deployment script
# Cross-compiles, uploads binaries, and starts services on the remote EC2 instance.
#
# Required environment variables:
#   DEPLOY_HOST   - EC2 IP or hostname (e.g. 1.2.3.4)
#   DEPLOY_KEY    - Path to SSH private key (e.g. ~/.ssh/my-key.pem)
#   DEPLOY_DOMAIN - Public domain name (e.g. demo.owulveryck.info)
#
# Optional:
#   DEPLOY_USER   - SSH user (default: ubuntu)
#   ARENA_PORT    - Arena port on remote (default: 8888)
#   GRAPH_PORT    - Shopping Graph port on remote (default: 9000)
#   OBS_PORT      - Obs Hub port on remote (default: 9002)
#   COST_PRICE    - Cost price in cents (default: 5000)
#   PRODUCT_NAME  - Product name (default: "Casque Audio")
set -e

# Validate required env vars
for var in DEPLOY_HOST DEPLOY_KEY DEPLOY_DOMAIN; do
  if [ -z "${!var}" ]; then
    echo "ERROR: $var is not set"
    echo "Usage: DEPLOY_HOST=1.2.3.4 DEPLOY_KEY=~/.ssh/key.pem DEPLOY_DOMAIN=demo.owulveryck.info $0"
    exit 1
  fi
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

DEPLOY_USER=${DEPLOY_USER:-ubuntu}
ARENA_PORT=${ARENA_PORT:-8888}
GRAPH_PORT=${GRAPH_PORT:-9000}
OBS_PORT=${OBS_PORT:-9002}
COST_PRICE=${COST_PRICE:-5000}
PRODUCT_NAME=${PRODUCT_NAME:-"Casque Audio"}

SSH_OPTS="-i $DEPLOY_KEY -o StrictHostKeyChecking=no -o ConnectTimeout=10"
SSH_CMD="ssh $SSH_OPTS $DEPLOY_USER@$DEPLOY_HOST"
SCP_CMD="scp $SSH_OPTS"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Step 1: Cross-compile for linux/amd64
echo -e "${BLUE}Cross-compiling for linux/amd64...${NC}"
cd "$ROOT_DIR"
mkdir -p demo/bin/linux

GOOS=linux GOARCH=amd64 go build -o demo/bin/linux/shopping-graph ./demo/cmd/shopping-graph/
GOOS=linux GOARCH=amd64 go build -o demo/bin/linux/obs-hub ./demo/cmd/obs-hub/
GOOS=linux GOARCH=amd64 go build -o demo/bin/linux/arena ./demo/cmd/arena/
echo -e "${GREEN}Build complete.${NC}"

# Step 2: Generate Caddyfile from template
echo -e "${BLUE}Generating Caddyfile...${NC}"
sed "s/__DOMAIN__/$DEPLOY_DOMAIN/g" demo/scripts/Caddyfile.tpl > demo/bin/linux/Caddyfile

# Step 3: Upload binaries and Caddyfile
echo -e "${BLUE}Uploading to $DEPLOY_HOST...${NC}"
$SSH_CMD "mkdir -p /opt/demo"
$SCP_CMD demo/bin/linux/shopping-graph demo/bin/linux/obs-hub demo/bin/linux/arena demo/bin/linux/Caddyfile "$DEPLOY_USER@$DEPLOY_HOST:/opt/demo/"
echo -e "${GREEN}Upload complete.${NC}"

# Step 4: Stop existing services and start new ones
echo -e "${BLUE}Starting services on remote...${NC}"
$SSH_CMD bash <<REMOTE
set -e

# Stop existing demo processes
pkill -f '/opt/demo/shopping-graph' 2>/dev/null || true
pkill -f '/opt/demo/obs-hub' 2>/dev/null || true
pkill -f '/opt/demo/arena' 2>/dev/null || true
sleep 1

cd /opt/demo
chmod +x shopping-graph obs-hub arena

# Start Shopping Graph
nohup ./shopping-graph \
  --port $GRAPH_PORT \
  --dynamic \
  --obs-url "http://localhost:$OBS_PORT" \
  --poll-interval 10s \
  > shopping-graph.log 2>&1 &
sleep 1

# Start Obs Hub
nohup ./obs-hub \
  --port $OBS_PORT \
  --graph-url "http://localhost:$GRAPH_PORT" \
  --arena-url "http://localhost:$ARENA_PORT" \
  > obs-hub.log 2>&1 &
sleep 1

# Start Arena with public base URL
nohup ./arena \
  --port $ARENA_PORT \
  --base-url "https://$DEPLOY_DOMAIN" \
  --cost-price $COST_PRICE \
  --product-name "$PRODUCT_NAME" \
  --graph-url "http://localhost:$GRAPH_PORT" \
  --obs-url "http://localhost:$OBS_PORT" \
  > arena.log 2>&1 &
sleep 1

# Update Caddy configuration
sudo cp /opt/demo/Caddyfile /etc/caddy/Caddyfile
sudo systemctl reload caddy

echo "Services started:"
pgrep -la 'shopping-graph|obs-hub|arena' || echo "(check logs if no processes found)"
REMOTE

echo ""
echo -e "${GREEN}=== Deployment complete ===${NC}"
echo ""
echo -e "  Arena:          ${BLUE}https://$DEPLOY_DOMAIN/${NC}"
echo -e "  Arena (auto):   ${BLUE}https://$DEPLOY_DOMAIN/auto${NC}"
echo -e "  Obs Hub:        ${BLUE}https://obs.$DEPLOY_DOMAIN/arena${NC}"
echo -e "  Shopping Graph: ${BLUE}https://graph.$DEPLOY_DOMAIN/health${NC}"
echo ""
echo -e "${GREEN}To start the client agent locally:${NC}"
echo -e "  demo/bin/client \\"
echo -e "    --graph-url https://graph.$DEPLOY_DOMAIN \\"
echo -e "    --obs-url https://obs.$DEPLOY_DOMAIN"
echo ""
echo -e "${GREEN}To check remote logs:${NC}"
echo -e "  $SSH_CMD 'tail -f /opt/demo/*.log'"
echo ""
echo -e "${GREEN}To stop remote services:${NC}"
echo -e "  $SSH_CMD 'pkill -f /opt/demo/'"
