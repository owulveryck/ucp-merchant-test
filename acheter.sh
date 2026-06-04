#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  AGENT ACHETEUR - COMMANDER UN PRODUIT                  ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Demander ce que l'utilisateur veut acheter
echo -e "${BLUE}Que veux-tu acheter ?${NC}"
echo ""
echo "1. Casque audio (budget \$100)"
echo "2. Laptop (budget \$1000)"
echo "3. Souris gaming (budget \$80)"
echo "4. Autre (personnalisé)"
echo ""
echo -n "Choix (1-4) : "
read CHOICE

case $CHOICE in
    1)
        QUERY="I want to buy headphones"
        BUDGET=100
        ;;
    2)
        QUERY="I want to buy a laptop"
        BUDGET=1000
        ;;
    3)
        QUERY="I want to buy a gaming mouse"
        BUDGET=80
        ;;
    4)
        echo -n "Que veux-tu acheter ? : "
        read QUERY
        echo -n "Budget (en \$) : "
        read BUDGET
        ;;
    *)
        echo "Choix invalide"
        exit 1
        ;;
esac

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🛒 Envoi de la commande à l'agent acheteur...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Demande : $QUERY"
echo "Budget  : \$$BUDGET"
echo ""

# Chercher le produit dans le Shopping Graph
echo "🔍 Recherche du produit dans le Shopping Graph..."
echo ""

SEARCH_RESULT=$(curl -s -X POST http://localhost:9000/search \
  -H "Content-Type: application/json" \
  -d "{\"query\": \"casque\", \"limit\": 10}")

echo "$SEARCH_RESULT" | jq -r '.results[] | "  • \(.merchant_name): $\(.price/100) - \(.in_stock | if . then "✅ En stock" else "❌ Rupture" end)"' 2>/dev/null || echo "Résultats de recherche disponibles"

echo ""
echo -e "${BLUE}📊 Qui a le meilleur prix ?${NC}"
echo ""

# Trouver le moins cher
CHEAPEST=$(echo "$SEARCH_RESULT" | jq -r '[.results[] | select(.in_stock == true)] | sort_by(.price) | .[0] | {merchant: .merchant_name, price: .price, id: .merchant_id}')

if [ -z "$CHEAPEST" ] || [ "$CHEAPEST" = "null" ]; then
    echo -e "${RED}❌ Aucun marchand trouvé${NC}"
    echo "Assure-toi d'avoir créé des marchands sur http://localhost:8888"
    exit 1
fi

WINNER_NAME=$(echo "$CHEAPEST" | jq -r '.merchant')
WINNER_PRICE=$(echo "$CHEAPEST" | jq -r '.price')
WINNER_ID=$(echo "$CHEAPEST" | jq -r '.id')

echo -e "${GREEN}🏆 GAGNANT: $WINNER_NAME${NC}"
echo -e "   Prix: ${YELLOW}\$$(echo "scale=2; $WINNER_PRICE/100" | bc)${NC}"
echo ""

# Simuler un achat
echo -e "${BLUE}🛒 Achat en cours...${NC}"
sleep 1

# Créer un checkout avec AUTO_COMPETE
CHECKOUT=$(curl -s -X POST "http://localhost:8888/$WINNER_ID/checkouts" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "casque_audio", "quantity": 1}],
    "customer": {"email": "agent@acheteur.com"},
    "discount_codes": ["AUTO_COMPETE"]
  }')

FINAL_PRICE=$(echo "$CHECKOUT" | jq -r '.totals[] | select(.type == "total") | .amount' 2>/dev/null)

if [ -n "$FINAL_PRICE" ] && [ "$FINAL_PRICE" != "null" ]; then
    echo ""
    echo -e "${GREEN}✅ ACHAT RÉUSSI !${NC}"
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "  Marchand gagnant : ${GREEN}$WINNER_NAME${NC}"
    echo -e "  Prix final : ${YELLOW}\$$(echo "scale=2; $FINAL_PRICE/100" | bc)${NC}"
    echo -e "  Code AUTO_COMPETE utilisé : ${GREEN}✅${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}📊 Le système 3-agents a calculé le meilleur prix !${NC}"
else
    echo -e "${RED}❌ Erreur lors de l'achat${NC}"
    echo "$CHECKOUT" | jq '.' 2>/dev/null || echo "$CHECKOUT"
fi

echo ""
echo -e "${BLUE}👀 Regarde l'arène pour voir la compétition :${NC}"
echo "   ${YELLOW}http://localhost:9002/arena${NC}"
echo ""
