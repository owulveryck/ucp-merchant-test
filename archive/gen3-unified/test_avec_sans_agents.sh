#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  COMPARAISON: AVEC vs SANS SYSTÈME MULTI-AGENTS         ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Trouver un marchand
TENANT_ID=$(cat logs/arena.log 2>/dev/null | grep "registered tenant" | head -1 | grep -o '([a-f0-9]\{8\})' | tr -d '()' || echo "")

if [ -z "$TENANT_ID" ]; then
    echo -e "${RED}❌ Aucun marchand trouvé${NC}"
    echo "Lance d'abord: ./run_full_demo.sh"
    echo "Puis crée un marchand sur http://localhost:8888"
    exit 1
fi

TENANT_NAME=$(cat logs/arena.log 2>/dev/null | grep "registered tenant" | head -1 | sed 's/.*registered tenant \([^ ]*\) .*/\1/')

echo -e "${BLUE}🏪 Marchand sélectionné: $TENANT_NAME${NC}"
echo "   Dashboard: http://localhost:8888/$TENANT_ID/dashboard"
echo ""

# Récupérer le prix actuel
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}TEST 1: SANS SYSTÈME MULTI-AGENTS (pricing manuel)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

CONFIG=$(curl -s "http://localhost:8888/$TENANT_ID/api/config")
PRICE=$(echo "$CONFIG" | jq -r '.selling_price // 7000')

echo "Prix configuré : \$$(echo "scale=2; $PRICE/100" | bc)"
echo ""
echo -e "${GREEN}→ Pas de réduction automatique${NC}"
echo -e "${GREEN}→ Prix manuel défini par le marchand${NC}"
echo ""

sleep 2

# Test AVEC le système multi-agents
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}TEST 2: AVEC SYSTÈME MULTI-AGENTS (code AUTO_COMPETE)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

RESPONSE=$(curl -s -X POST "http://localhost:8888/$TENANT_ID/api/test-auto-compete")

if echo "$RESPONSE" | grep -q "error"; then
    echo -e "${RED}❌ Système multi-agents non disponible${NC}"
    echo "$RESPONSE"
    exit 1
fi

echo "$RESPONSE" | jq -r '
"🤖 LES 4 AGENTS ONT TRAVAILLÉ:",
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
"💰 RÉSULTAT FINAL",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   Prix manuel :     $" + ((.current_price / 100) | tostring),
"   Réduction auto :  -$" + ((.discount_amount / 100) | tostring),
"   Prix optimisé :   $" + ((.final_price / 100) | tostring) + " ✨",
"   Marge finale :    " + (.margin_percent | tostring) + "%",
""
'

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}📌 RÉSUMÉ${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${RED}SANS AUTO_COMPETE :${NC}"
echo "  → Prix manuel fixe"
echo "  → Pas d'analyse de marché"
echo "  → Pas d'optimisation"
echo ""
echo -e "${GREEN}AVEC AUTO_COMPETE :${NC}"
echo "  → 4 agents analysent le marché"
echo "  → Prix optimisé automatiquement"
echo "  → Meilleure compétitivité"
echo ""
echo -e "${BLUE}🎯 Pour déclencher le système multi-agents :${NC}"
echo "  Utilise le code promo ${YELLOW}AUTO_COMPETE${NC} lors du checkout"
echo ""
echo -e "${BLUE}🌐 Dashboard marchand :${NC}"
echo "  http://localhost:8888/$TENANT_ID/dashboard"
echo ""
