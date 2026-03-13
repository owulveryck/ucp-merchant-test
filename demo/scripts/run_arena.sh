#!/bin/bash
# Arena demo launcher
# Starts: Shopping Graph (dynamic) + Obs Hub + Arena Server + optional Client Agent
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Ports
ARENA_PORT=${ARENA_PORT:-8888}
GRAPH_PORT=${GRAPH_PORT:-9000}
OBS_PORT=${OBS_PORT:-9002}
COST_PRICE=${COST_PRICE:-5000}
PRODUCT_NAME=${PRODUCT_NAME:-"Casque Audio"}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

cleanup() {
    echo -e "\n${RED}Shutting down...${NC}"
    kill $GRAPH_PID $OBS_PID $ARENA_PID $CLIENT_PID 2>/dev/null || true
    wait 2>/dev/null || true
    echo -e "${GREEN}Done.${NC}"
}
trap cleanup EXIT

# Build
echo -e "${BLUE}Building...${NC}"
cd "$ROOT_DIR"
go build -o demo/bin/shopping-graph ./demo/cmd/shopping-graph/
go build -o demo/bin/obs-hub ./demo/cmd/obs-hub/
go build -o demo/bin/arena ./demo/cmd/arena/
go build -o demo/bin/client ./demo/cmd/client/ 2>/dev/null || echo "Client build skipped (may need genai dependency)"

# Start Shopping Graph (dynamic mode)
echo -e "${GREEN}Starting Shopping Graph on port $GRAPH_PORT (dynamic mode)...${NC}"
demo/bin/shopping-graph \
    --port "$GRAPH_PORT" \
    --dynamic \
    --obs-url "http://localhost:$OBS_PORT" \
    --poll-interval 10s &
GRAPH_PID=$!
sleep 1

# Start Obs Hub
echo -e "${GREEN}Starting Obs Hub on port $OBS_PORT...${NC}"
demo/bin/obs-hub \
    --port "$OBS_PORT" \
    --graph-url "http://localhost:$GRAPH_PORT" \
    --arena-url "http://localhost:$ARENA_PORT" &
OBS_PID=$!
sleep 1

# Start Arena
echo -e "${GREEN}Starting Arena on port $ARENA_PORT...${NC}"
demo/bin/arena \
    --port "$ARENA_PORT" \
    --cost-price "$COST_PRICE" \
    --product-name "$PRODUCT_NAME" \
    --graph-url "http://localhost:$GRAPH_PORT" \
    --obs-url "http://localhost:$OBS_PORT" &
ARENA_PID=$!
sleep 1

# Start Client Agent (if GCP project set)
CLIENT_PID=""
if [ -n "$GOOGLE_CLOUD_PROJECT" ] && [ -f demo/bin/client ]; then
    echo -e "${GREEN}Starting Client Agent...${NC}"
    demo/bin/client \
        --graph-url "http://localhost:$GRAPH_PORT" \
        --obs-url "http://localhost:$OBS_PORT" &
    CLIENT_PID=$!
else
    echo -e "${BLUE}Client Agent skipped (set GOOGLE_CLOUD_PROJECT to enable)${NC}"
fi

echo ""
echo -e "${GREEN}=== Arena is ready ===${NC}"
echo -e "  Landing page:  ${BLUE}http://localhost:$ARENA_PORT/${NC}"
echo -e "  Arena Monitor: ${BLUE}http://localhost:$OBS_PORT/arena${NC}"
echo -e "  Shopping Graph: ${BLUE}http://localhost:$GRAPH_PORT/health${NC}"
echo -e "  Obs Hub:       ${BLUE}http://localhost:$OBS_PORT/${NC}"
echo ""
echo "Press Ctrl+C to stop."

wait
