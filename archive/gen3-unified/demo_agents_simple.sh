#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  DÉMO SYSTÈME MULTI-AGENTS - VERSION SIMPLE             ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Trouver le premier marchand
TENANT_ID=$(cat logs/arena.log 2>/dev/null | grep "registered tenant" | grep "MULTI-AGENT" | head -1 | grep -o '([a-f0-9]\{8\})' | tr -d '()' || echo "")

if [ -z "$TENANT_ID" ]; then
    echo -e "${RED}❌ Aucun marchand avec competitive pricing trouvé${NC}"
    echo ""
    echo "As-tu bien lancé: ./run_full_demo.sh ?"
    echo ""
    exit 1
fi

TENANT_NAME=$(cat logs/arena.log | grep "$TENANT_ID" | grep "registered tenant" | head -1 | sed 's/.*registered tenant \([^ ]*\) .*/\1/')

echo -e "${BLUE}🏪 Marchand sélectionné: $TENANT_NAME${NC}"
echo "   ID: $TENANT_ID"
echo "   Dashboard: http://localhost:8888/$TENANT_ID/dashboard"
echo ""

sleep 1

echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 APPEL DU SYSTÈME MULTI-AGENTS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Le code promo AUTO_COMPETE déclenche les 4 agents..."
echo ""

sleep 1

RESPONSE=$(curl -s -X POST "http://localhost:8888/$TENANT_ID/api/test-auto-compete")

if echo "$RESPONSE" | grep -q "error"; then
    echo -e "${RED}❌ Erreur du système:${NC}"
    echo "$RESPONSE" | jq '.'
    echo ""
    echo "Vérifie que competitive pricing est activé dans les logs:"
    echo "  cat logs/arena.log | grep 'MULTI-AGENT.*$TENANT_NAME'"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ SYSTÈME MULTI-AGENTS ACTIVÉ !${NC}"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Afficher les décisions des agents
echo "$RESPONSE" | jq -r '
"",
"🤖 LES 4 AGENTS ONT ANALYSÉ LE MARCHÉ:",
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"📊 AGENT 1: PRICE INTELLIGENCE",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   " + .reasoning.agent1,
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"📈 AGENT 2: MARKET ANALYSIS",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   " + .reasoning.agent2,
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"🎯 AGENT 3: STRATEGY RECOMMENDER",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   " + (.reasoning.agent3 | gsub("<br>"; "\n   ") | gsub("<strong>"; "*** ") | gsub("</strong>"; " ***")),
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"✅ AGENT 4: MARGIN VALIDATOR",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   " + .reasoning.agent4,
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"💰 DÉCISION FINALE DES AGENTS",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   Prix configuré par marchand :  $" + ((.current_price / 100) | tostring),
"   Réduction calculée (agents) :  -$" + ((.discount_amount / 100) | tostring),
"   ───────────────────────────────────────────────",
"   Prix final optimisé :          $" + ((.final_price / 100) | tostring) + " ✨",
"   Marge finale :                 " + (.margin_percent | tostring) + "%",
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
'

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}📌 COMMENT ÇA MARCHE${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "1️⃣  Le client utilise le code promo: ${YELLOW}AUTO_COMPETE${NC}"
echo ""
echo "2️⃣  Les 4 agents s'activent automatiquement:"
echo "    • Agent 1: Cherche les prix concurrents"
echo "    • Agent 2: Analyse la position marché"
echo "    • Agent 3: Recommande une stratégie"
echo "    • Agent 4: Valide les marges"
echo ""
echo "3️⃣  Le système calcule le MEILLEUR PRIX pour battre la concurrence"
echo ""
echo "4️⃣  Le client obtient automatiquement ce prix optimisé !"
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}🌐 Dashboard du marchand:${NC}"
echo "   http://localhost:8888/$TENANT_ID/dashboard"
echo ""
echo -e "${BLUE}📜 Voir les logs détaillés en temps réel:${NC}"
echo "   tail -f logs/arena.log | grep -E 'Agent|Orchestrator|AUTO_COMPETE'"
echo ""
echo -e "${BLUE}🔄 Relancer le test:${NC}"
echo "   ./demo_agents_simple.sh"
echo ""
