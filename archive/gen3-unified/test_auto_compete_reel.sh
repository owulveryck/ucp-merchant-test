#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  TEST RÉEL DU SYSTÈME MULTI-AGENTS AVEC AUTO_COMPETE    ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Trouver un marchand
TENANT_ID=$(cat logs/arena.log 2>/dev/null | grep "registered tenant" | head -1 | grep -o '([a-f0-9]\{8\})' | tr -d '()' || echo "")

if [ -z "$TENANT_ID" ]; then
    echo -e "${RED}❌ Aucun marchand trouvé${NC}"
    echo "Lance d'abord: ./run_full_demo.sh"
    exit 1
fi

TENANT_NAME=$(cat logs/arena.log | grep "registered tenant" | head -1 | sed 's/.*registered tenant \([^ ]*\) .*/\1/')

echo -e "${BLUE}🏪 Marchand: $TENANT_NAME (ID: $TENANT_ID)${NC}"
echo "   Dashboard: http://localhost:8888/$TENANT_ID/dashboard"
echo ""

# Récupérer l'URL de découverte UCP
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}ÉTAPE 1: Découverte UCP${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

UCP_DISCOVERY=$(curl -s "http://localhost:8888/$TENANT_ID/.well-known/ucp")
CHECKOUT_URL=$(echo "$UCP_DISCOVERY" | jq -r '.endpoints.checkouts')

echo "✓ UCP endpoint trouvé: $CHECKOUT_URL"
echo ""

# Créer un panier
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}ÉTAPE 2: Création du panier${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

CART_RESPONSE=$(curl -s -X POST "http://localhost:8888/$TENANT_ID/carts" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": "casque_audio",
        "quantity": 1
      }
    ]
  }')

CART_ID=$(echo "$CART_RESPONSE" | jq -r '.id // .cart_id // empty')

if [ -z "$CART_ID" ]; then
    echo -e "${RED}❌ Impossible de créer le panier${NC}"
    echo "$CART_RESPONSE"
    exit 1
fi

echo "✓ Panier créé: $CART_ID"
echo ""

# TEST 1: Checkout SANS AUTO_COMPETE
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}TEST 1: Checkout SANS AUTO_COMPETE${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

CHECKOUT1=$(curl -s -X POST "$CHECKOUT_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"cart_id\": \"$CART_ID\",
    \"customer\": {
      \"email\": \"test@example.com\"
    }
  }")

PRICE1=$(echo "$CHECKOUT1" | jq -r '.totals[] | select(.type == "total") | .amount')

echo "Prix SANS agents: \$$(echo "scale=2; $PRICE1/100" | bc)"
echo -e "${YELLOW}→ Pas de réduction automatique${NC}"
echo ""

sleep 2

# TEST 2: Checkout AVEC AUTO_COMPETE
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}TEST 2: Checkout AVEC AUTO_COMPETE ✨${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo "📝 Envoi du code promo AUTO_COMPETE..."
echo ""

CHECKOUT2=$(curl -s -X POST "$CHECKOUT_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"cart_id\": \"$CART_ID\",
    \"customer\": {
      \"email\": \"test@example.com\"
    },
    \"discount_codes\": [\"AUTO_COMPETE\"]
  }")

PRICE2=$(echo "$CHECKOUT2" | jq -r '.totals[] | select(.type == "total") | .amount')
DISCOUNT=$(echo "$CHECKOUT2" | jq -r '.totals[] | select(.type == "discount") | .amount // 0')

echo "🤖 LE SYSTÈME MULTI-AGENTS A TRAVAILLÉ !"
echo ""
echo "Prix initial:     \$$(echo "scale=2; $PRICE1/100" | bc)"
echo "Réduction agents: -\$$(echo "scale=2; $DISCOUNT/100" | bc)"
echo "Prix final:       \$$(echo "scale=2; $PRICE2/100" | bc) ✨"
echo ""

# Afficher les détails des totaux
echo "Détails du checkout:"
echo "$CHECKOUT2" | jq -r '.totals[] | "  - \(.type): $\(.amount / 100)"'

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}📊 RÉSUMÉ${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${RED}SANS AUTO_COMPETE:${NC}"
echo "  Prix: \$$(echo "scale=2; $PRICE1/100" | bc)"
echo "  Aucune analyse, prix fixe"
echo ""
echo -e "${GREEN}AVEC AUTO_COMPETE:${NC}"
echo "  Prix: \$$(echo "scale=2; $PRICE2/100" | bc)"
echo "  Économie: \$$(echo "scale=2; ($PRICE1-$PRICE2)/100" | bc)"
echo "  4 agents ont analysé le marché !"
echo ""
echo -e "${BLUE}📜 Voir les logs des agents:${NC}"
echo "  tail -f logs/arena.log | grep -E 'Agent|Orchestrator'"
echo ""
echo -e "${BLUE}🎯 Le code magique: ${YELLOW}AUTO_COMPETE${NC}"
echo ""
