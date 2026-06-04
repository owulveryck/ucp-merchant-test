#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  🏆 ARENA CHALLENGE - SYSTÈME MULTI-AGENTS              ║"
echo "║  Scénario : Tu rentres perdant, tu finis gagnant !      ║"
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

export GOOGLE_CLOUD_PROJECT="bsjxygz-gcp-octo-lille"

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
echo -e "${BLUE}🚀 Démarrage de l'arène...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 1. Shopping Graph
echo -e "${BLUE}[1/3]${NC} Shopping Graph (port 9000)..."
bin/shopping-graph --port 9000 --dynamic --poll-interval 10s > logs/shopping-graph.log 2>&1 &
SHOPPING_PID=$!
sleep 2

# 2. Observability Hub
echo -e "${BLUE}[2/3]${NC} Observability Hub (port 9002)..."
bin/obs-hub --port 9002 --graph-url http://localhost:9000 --arena-url http://localhost:8888 > logs/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2

# 3. Arena avec système 3-agents activé
echo -e "${BLUE}[3/3]${NC} Arena Merchant (port 8888) ${YELLOW}avec SYSTÈME MULTI-AGENTS${NC}..."
bin/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 --competitive-pricing --min-margin 10 > logs/arena.log 2>&1 &
ARENA_PID=$!
sleep 3

# Vérifier que les services sont bien démarrés
if ! ps -p $SHOPPING_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Shopping Graph a crashé${NC}"
    cat logs/shopping-graph.log
    exit 1
fi
if ! ps -p $OBS_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Obs Hub a crashé${NC}"
    cat logs/obs-hub.log
    exit 1
fi
if ! ps -p $ARENA_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Arena a crashé${NC}"
    cat logs/arena.log
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Services opérationnels !${NC}"
echo ""

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${MAGENTA}🎮 CRÉATION DE L'ARÈNE DE COMPÉTITION${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Créer 4 marchands concurrents avec des prix agressifs
echo -e "${BLUE}📍 Création de 4 marchands concurrents...${NC}"
echo ""

COMPETITORS=(
    "MegaStore:6200:mega_001"
    "PrixCassés:5800:prix_002"
    "SuperDeals:6000:super_003"
    "TopPrix:5900:top_004"
)

for merchant in "${COMPETITORS[@]}"; do
    IFS=':' read -r name price id <<< "$merchant"

    # Inscription du marchand
    curl -s -X POST "http://localhost:8888/register" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$name\", \"email\": \"${id}@competitor.com\"}" \
        -o /dev/null

    # Configurer le prix de base
    curl -s -X PUT "http://localhost:8888/${id}/api/config" \
        -H "Content-Type: application/json" \
        -d "{\"base_price\": $price}" \
        -o /dev/null

    echo -e "  ${GREEN}✓${NC} $name configuré à \$$(echo "scale=2; $price/100" | bc)"
    sleep 0.5
done

echo ""
echo -e "${GREEN}✅ 4 concurrents créés${NC}"
echo ""

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🎯 DÉMO EN 3 ÉTAPES${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo -e "${BLUE}1. Crée ton marchand (sans optimiser)${NC}"
echo "   ${GREEN}→ http://localhost:8888${NC}"
echo "   Laisse le prix par défaut (~\$70) → tu seras le plus cher"
echo ""

echo -e "${BLUE}2. Teste l'achat dans l'arène${NC}"
echo "   ${GREEN}→ http://localhost:9002/arena${NC}"
echo "   Tape \"Achète un casque\" → un concurrent gagne"
echo ""

echo -e "${BLUE}3. Active le système 3-agents${NC}"
echo "   Dashboard → \"💡 Calculer meilleur prix\" → Applique"
echo "   Retour arène → \"Achète un casque\" → ${GREEN}TU GAGNES !${NC}"
echo ""

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}⚡ Services actifs${NC} | Ctrl+C pour arrêter"
echo ""

# Attendre
wait
