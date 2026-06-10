#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  DÉMO COMPLÈTE - SYSTÈME MULTI-AGENTS                   ║"
echo "║  Tout-en-un : Lancement + Test automatique              ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Fonction de nettoyage
cleanup() {
    echo ""
    echo -e "${RED}🛑 Arrêt des services...${NC}"
    kill $SHOPPING_PID $OBS_PID $ARENA_PID 2>/dev/null || true
    killall shopping-graph obs-hub arena 2>/dev/null || true
    sleep 1
    echo -e "${GREEN}✓ Services arrêtés${NC}"
}

trap cleanup EXIT

# Configuration
export GOOGLE_CLOUD_PROJECT="bsjxygz-gcp-octo-lille"

# Créer les dossiers
mkdir -p logs bin

echo -e "${YELLOW}🔨 Build...${NC}"
cd demo
go build -o ../bin/shopping-graph ./cmd/shopping-graph 2>/dev/null
go build -o ../bin/obs-hub ./cmd/obs-hub 2>/dev/null
go build -o ../bin/arena ./cmd/arena 2>/dev/null
cd ..
echo -e "${GREEN}✓ Build terminé${NC}"
echo ""

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🚀 Démarrage des services...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 1. Shopping Graph
echo -e "${BLUE}[1/3]${NC} Shopping Graph (port 9000)..."
bin/shopping-graph --port 9000 --dynamic --poll-interval 10s > logs/shopping-graph.log 2>&1 &
SHOPPING_PID=$!
sleep 2
if ps -p $SHOPPING_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    exit 1
fi

# 2. Observability Hub
echo -e "${BLUE}[2/3]${NC} Observability Hub (port 9002)..."
bin/obs-hub --port 9002 --graph-url http://localhost:9000 --arena-url http://localhost:8888 > logs/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2
if ps -p $OBS_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    exit 1
fi

# 3. Arena avec MULTI-AGENTS activé
echo -e "${BLUE}[3/3]${NC} Arena Merchant (port 8888) ${YELLOW}avec SYSTÈME MULTI-AGENTS${NC}..."
bin/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 --competitive-pricing --min-margin 10 > logs/arena.log 2>&1 &
ARENA_PID=$!
sleep 6
if ps -p $ARENA_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré avec competitive pricing${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    cat logs/arena.log
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Services opérationnels !${NC}"
echo ""

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ SERVICES ACTIFS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}1. Crée tes marchands${NC}"
echo "   ${GREEN}→ http://localhost:8888${NC}"
echo ""
echo -e "${BLUE}2. Optimise avec le système 3-agents${NC}"
echo "   Dashboard → \"💡 Calculer meilleur prix\" → Applique"
echo ""
echo -e "${BLUE}3. Teste dans l'arène${NC}"
echo "   ${GREEN}→ http://localhost:9002/arena${NC}"
echo "   Tape \"Achète un casque\""
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}⚡ Ctrl+C pour arrêter${NC}"
echo ""

# Attendre que les processus se terminent (ou Ctrl+C)
wait
