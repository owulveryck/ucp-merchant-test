#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  DÉMONSTRATION COMPLÈTE SYSTÈME MULTI-AGENTS            ║"
echo "║  Lancement + Test automatique                           ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Fonction de nettoyage
cleanup() {
    echo ""
    echo -e "${RED}🛑 Arrêt des services...${NC}"
    [ -n "$SHOPPING_PID" ] && kill $SHOPPING_PID 2>/dev/null || true
    [ -n "$OBS_PID" ] && kill $OBS_PID 2>/dev/null || true
    [ -n "$ARENA_PID" ] && kill $ARENA_PID 2>/dev/null || true
    sleep 1
    echo -e "${GREEN}✓ Services arrêtés${NC}"
}

trap cleanup EXIT

# Configuration
export GOOGLE_CLOUD_PROJECT="bsjxygz-gcp-octo-lille"
SHOPPING_GRAPH_PORT=9000
OBS_HUB_PORT=9002
ARENA_PORT=8888

echo -e "${BLUE}📦 Build des composants...${NC}"
cd demo
go build -o bin/shopping-graph ./cmd/shopping-graph
go build -o bin/obs-hub ./cmd/obs-hub
go build -o bin/arena ./cmd/arena
cd ..
echo -e "${GREEN}✓ Build terminé${NC}"
echo ""

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🚀 Démarrage des services...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 1. Shopping Graph
echo -e "${BLUE}[1/3] Shopping Graph (port $SHOPPING_GRAPH_PORT)...${NC}"
cd demo
./bin/shopping-graph --port $SHOPPING_GRAPH_PORT > /tmp/shopping-graph.log 2>&1 &
SHOPPING_PID=$!
sleep 2
if ps -p $SHOPPING_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré (PID: $SHOPPING_PID)${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    exit 1
fi

# 2. Observability Hub
echo -e "${BLUE}[2/3] Observability Hub (port $OBS_HUB_PORT)...${NC}"
./bin/obs-hub --port $OBS_HUB_PORT > /tmp/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2
if ps -p $OBS_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré (PID: $OBS_PID)${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    exit 1
fi

# 3. Arena avec competitive pricing
echo -e "${BLUE}[3/3] Arena Merchant (port $ARENA_PORT) avec MULTI-AGENTS...${NC}"
./bin/arena --port $ARENA_PORT --competitive-pricing > /tmp/arena.log 2>&1 &
ARENA_PID=$!
sleep 6
if ps -p $ARENA_PID > /dev/null; then
    echo -e "${GREEN}      ✓ Démarré (PID: $ARENA_PID)${NC}"
else
    echo -e "${RED}      ✗ Échec${NC}"
    cat /tmp/arena.log
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Tous les services sont démarrés !${NC}"
echo ""

# Attendre qu'Arena soit complètement prêt
sleep 3

# Créer un marchand de test
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🏪 Création d'un marchand de test...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Attendre qu'un marchand soit enregistré (ils se créent automatiquement)
echo "Attente de l'enregistrement automatique des marchands..."
sleep 5

# Récupérer le premier marchand disponible
TENANT_ID=$(cat /tmp/arena.log | grep "registered tenant" | head -1 | grep -o '([a-f0-9]\{8\})' | tr -d '()')
TENANT_NAME=$(cat /tmp/arena.log | grep "registered tenant" | head -1 | sed 's/.*registered tenant \([^ ]*\) .*/\1/')

if [ -z "$TENANT_ID" ]; then
    echo -e "${RED}❌ Aucun marchand trouvé${NC}"
    echo ""
    echo "Crée un marchand manuellement :"
    echo "  1. Va sur http://localhost:$ARENA_PORT"
    echo "  2. Crée un marchand"
    echo ""
    echo "Appuie sur Ctrl+C pour arrêter les services"
    wait
    exit 1
fi

echo -e "${GREEN}✓ Marchand trouvé : $TENANT_NAME (ID: $TENANT_ID)${NC}"
echo ""

# Tester le système multi-agents
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 TEST DU SYSTÈME MULTI-AGENTS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Marchand : $TENANT_NAME"
echo "Dashboard : http://localhost:$ARENA_PORT/$TENANT_ID/dashboard"
echo ""

sleep 2

RESPONSE=$(curl -s -X POST "http://localhost:$ARENA_PORT/$TENANT_ID/api/test-auto-compete" 2>&1)

if echo "$RESPONSE" | grep -q "error"; then
    echo -e "${RED}❌ Erreur:${NC}"
    echo "$RESPONSE"
else
    echo "$RESPONSE" | jq -r '
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"📊 AGENT 1: PRICE INTELLIGENCE",
"   " + .reasoning.agent1,
"",
"📈 AGENT 2: MARKET ANALYSIS",
"   " + .reasoning.agent2,
"",
"🎯 AGENT 3: STRATEGY RECOMMENDER",
"   " + (.reasoning.agent3 | gsub("<br>"; "\n   ") | gsub("<strong>"; "") | gsub("</strong>"; "")),
"",
"✅ AGENT 4: MARGIN VALIDATOR",
"   " + .reasoning.agent4,
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"💰 DÉCISION FINALE",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   Prix initial :    $" + ((.current_price / 100) | tostring),
"   Réduction :       -$" + ((.discount_amount / 100) | tostring),
"   Prix final :      $" + ((.final_price / 100) | tostring) + " ✨",
"   Marge :           " + (.margin_percent | tostring) + "%",
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
' 2>/dev/null || echo "$RESPONSE"
fi

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📍 ACCÈS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "🌐 Dashboard :  http://localhost:$ARENA_PORT"
echo "🏪 Marchand :   http://localhost:$ARENA_PORT/$TENANT_ID/dashboard"
echo ""
echo "📊 Logs en temps réel :"
echo "   tail -f /tmp/arena.log | grep -E 'Agent|Orchestrator'"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}✨ Système opérationnel ! Appuie sur Ctrl+C pour arrêter.${NC}"
echo ""

# Attendre
wait
