#!/bin/bash
set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  DÉMO SYSTÈME MULTI-AGENTS UNIFIÉ                       ║"
echo "║  Agent Vendeur + Customer Growth + Compétitivité        ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Configuration
export GOOGLE_CLOUD_PROJECT="bsjxygz-gcp-octo-lille"
SHOPPING_GRAPH_PORT=9000
OBS_HUB_PORT=9002
ARENA_PORT=8888

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📦 Building all components...${NC}"
cd demo
go build -o bin/shopping-graph ./cmd/shopping-graph
go build -o bin/obs-hub ./cmd/obs-hub
go build -o bin/arena ./cmd/arena
go build -o bin/client ./cmd/client
cd ..

echo ""
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Fonction de nettoyage
cleanup() {
    echo ""
    echo -e "${RED}🛑 Stopping all services...${NC}"
    [ -n "$SHOPPING_PID" ] && kill $SHOPPING_PID 2>/dev/null || true
    [ -n "$OBS_PID" ] && kill $OBS_PID 2>/dev/null || true
    [ -n "$ARENA_PID" ] && kill $ARENA_PID 2>/dev/null || true
    sleep 1
    echo -e "${GREEN}✓ All services stopped${NC}"
}

trap cleanup EXIT

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🚀 Starting services (3/4 - Client is optional)...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 1. Shopping Graph
echo -e "${BLUE}[1/4] Starting Shopping Graph on port ${SHOPPING_GRAPH_PORT}...${NC}"
cd demo
./bin/shopping-graph --port $SHOPPING_GRAPH_PORT > /tmp/shopping-graph.log 2>&1 &
SHOPPING_PID=$!
sleep 2
if ps -p $SHOPPING_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Shopping Graph started (PID: $SHOPPING_PID)${NC}"
else
    echo -e "${RED}      ✗ Shopping Graph failed to start${NC}"
    exit 1
fi

# 2. Observability Hub
echo -e "${BLUE}[2/4] Starting Observability Hub on port ${OBS_HUB_PORT}...${NC}"
./bin/obs-hub --port $OBS_HUB_PORT > /tmp/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2
if ps -p $OBS_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Observability Hub started (PID: $OBS_PID)${NC}"
else
    echo -e "${RED}      ✗ Observability Hub failed to start${NC}"
    exit 1
fi

# 3. Arena (Merchant) with competitive pricing enabled
echo -e "${BLUE}[3/4] Starting Arena Merchant on port ${ARENA_PORT}...${NC}"
./bin/arena --port $ARENA_PORT --competitive-pricing > /tmp/arena.log 2>&1 &
ARENA_PID=$!
sleep 5
if ps -p $ARENA_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Arena Merchant started (PID: $ARENA_PID)${NC}"
else
    echo -e "${RED}      ✗ Arena Merchant failed to start${NC}"
    cat /tmp/arena.log
    exit 1
fi

# 4. Client Agent (optionnel - mode interactif)
echo -e "${BLUE}[4/4] Client Agent (mode interactif)...${NC}"
echo -e "${YELLOW}      ⓘ Client Agent disponible en mode CLI : ./demo/bin/client${NC}"
CLIENT_PID=""

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ All services are running!${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}📍 Services URLs:${NC}"
echo "   Shopping Graph:     http://localhost:$SHOPPING_GRAPH_PORT"
echo "   Observability Hub:  http://localhost:$OBS_HUB_PORT"
echo "   Arena Dashboard:    http://localhost:$ARENA_PORT"
echo "   Client Agent:       ./demo/bin/client (mode interactif, lance dans un autre terminal si besoin)"
echo ""
echo -e "${BLUE}🎯 DASHBOARD:${NC}"
echo -e "   ${GREEN}➜ http://localhost:$ARENA_PORT${NC}"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 VALEURS DE TEST - SYSTÈME MULTI-AGENTS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}1. CLIENT PREMIUM (garder à tout prix)${NC}"
echo "   Customer ID:  premium_vip_001"
echo "   Product:      casque_bluetooth"
echo "   Code promo:   AUTO_COMPETE"
echo "   Expected:     ✅ Réduction VIP 15% + prix compétitif"
echo ""
echo -e "${GREEN}2. CLIENT GOLD (important)${NC}"
echo "   Customer ID:  gold_customer_002"
echo "   Product:      laptop_pro"
echo "   Code promo:   AUTO_COMPETE"
echo "   Expected:     ✅ Réduction VIP 10% + prix compétitif"
echo ""
echo -e "${GREEN}3. CLIENT SILVER (bon)${NC}"
echo "   Customer ID:  silver_customer_003"
echo "   Product:      souris_gaming"
echo "   Code promo:   AUTO_COMPETE"
echo "   Expected:     ✅ Réduction VIP 5% + prix compétitif"
echo ""
echo -e "${YELLOW}4. CLIENT STANDARD (pas prioritaire)${NC}"
echo "   Customer ID:  standard_customer_999"
echo "   Product:      clavier_meca"
echo "   Code promo:   AUTO_COMPETE"
echo "   Expected:     ❌ Pas de bonus VIP, juste prix compétitif"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📝 COMMENT TESTER:${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "1. Ouvre le dashboard: http://localhost:$ARENA_PORT"
echo "2. Clique sur 'Test AUTO_COMPETE'"
echo "3. Le code AUTO_COMPETE déclenche le système multi-agents:"
echo "   → Agent 1 (Vendeur) coordonne"
echo "   → Agent 2 (Customer Growth) analyse le client"
echo "   → Agent 3 (Compétitivité) analyse le marché"
echo "   → Agent 1 décide du prix final"
echo ""
echo -e "${BLUE}📊 Logs en temps réel:${NC}"
echo "   Shopping Graph:  tail -f /tmp/shopping-graph.log"
echo "   Obs Hub:         tail -f /tmp/obs-hub.log"
echo "   Arena:           tail -f /tmp/arena.log"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}Press Ctrl+C to stop all services${NC}"
echo ""

# Attendre
wait
