#!/bin/bash
# Script de test de l'architecture multi-agents

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Test Architecture Multi-Agents - AUTO_COMPETE  ║${NC}"
echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo ""

# Vérifier que le Shopping Graph tourne
echo -e "${YELLOW}[1/6]${NC} Vérification Shopping Graph..."
if ! curl -s http://localhost:9000/search > /dev/null 2>&1; then
    echo -e "${RED}❌ Shopping Graph n'est pas accessible sur :9000${NC}"
    echo -e "${YELLOW}💡 Lancez d'abord :${NC}"
    echo -e "   cd demo && go run ./cmd/shopping-graph --port 9000"
    exit 1
fi
echo -e "${GREEN}✓ Shopping Graph OK${NC}"
echo ""

# Vérifier que les merchants tournent
echo -e "${YELLOW}[2/6]${NC} Vérification Merchants..."
MERCHANTS_OK=0
for PORT in 8182 8183 8184; do
    if curl -s http://localhost:$PORT/api/products > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Merchant sur :$PORT OK${NC}"
        ((MERCHANTS_OK++))
    else
        echo -e "${RED}❌ Merchant sur :$PORT non accessible${NC}"
    fi
done

if [ $MERCHANTS_OK -lt 2 ]; then
    echo -e "${RED}❌ Besoin d'au moins 2 merchants pour tester${NC}"
    echo -e "${YELLOW}💡 Lancez les merchants :${NC}"
    echo -e "   Terminal 1: go run ./sample_implementation --port 8182 --data-dir demo/data/merchant_a --merchant-name SuperShop"
    echo -e "   Terminal 2: go run ./sample_implementation --port 8183 --data-dir demo/data/merchant_b --merchant-name MegaMart"
    exit 1
fi
echo ""

# Récupérer les prix de base
echo -e "${YELLOW}[3/6]${NC} Récupération des prix concurrents..."

PRODUCT="prod_roses_bouquet"
SUPERSH0P_PRICE=$(curl -s http://localhost:8182/api/products | jq -r ".[] | select(.id==\"$PRODUCT\") | .price")
MEGAMART_PRICE=$(curl -s http://localhost:8183/api/products | jq -r ".[] | select(.id==\"$PRODUCT\") | .price")

if [ -z "$SUPERSH0P_PRICE" ] || [ -z "$MEGAMART_PRICE" ]; then
    echo -e "${RED}❌ Impossible de récupérer les prix${NC}"
    exit 1
fi

echo -e "  SuperShop (8182) : ${BLUE}\$$(echo "scale=2; $SUPERSH0P_PRICE/100" | bc)${NC}"
echo -e "  MegaMart  (8183) : ${BLUE}\$$(echo "scale=2; $MEGAMART_PRICE/100" | bc)${NC}"

if [ $SUPERSH0P_PRICE -lt $MEGAMART_PRICE ]; then
    echo -e "${YELLOW}⚠️  SuperShop est déjà moins cher que MegaMart${NC}"
    echo -e "${YELLOW}    Le test sera moins spectaculaire mais fonctionnera quand même${NC}"
fi
echo ""

# Attendre indexation Shopping Graph
echo -e "${YELLOW}[4/6]${NC} Vérification indexation Shopping Graph..."
INDEXED=$(curl -s -X POST http://localhost:9000/search \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"$PRODUCT\",\"limit\":10}" | jq '.total')

if [ "$INDEXED" -lt 2 ]; then
    echo -e "${YELLOW}⏳ En attente d'indexation (30s)...${NC}"
    sleep 30
    INDEXED=$(curl -s -X POST http://localhost:9000/search \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"$PRODUCT\",\"limit\":10}" | jq '.total')
fi

echo -e "${GREEN}✓ Shopping Graph a indexé $INDEXED merchants${NC}"
echo ""

# Créer checkout SANS AUTO_COMPETE
echo -e "${YELLOW}[5/6]${NC} Test 1: Checkout SANS AUTO_COMPETE..."

CHECKOUT_RESPONSE=$(curl -s -X POST http://localhost:8182/checkout \
    -H "Content-Type: application/json" \
    -d "{
        \"items\": [{
            \"product_id\": \"$PRODUCT\",
            \"quantity\": 1
        }]
    }")

CHECKOUT_ID=$(echo "$CHECKOUT_RESPONSE" | jq -r '.id')
TOTAL_BEFORE=$(echo "$CHECKOUT_RESPONSE" | jq -r '.totals[] | select(.type=="total") | .amount')

echo -e "  Checkout ID: ${BLUE}$CHECKOUT_ID${NC}"
echo -e "  Total SANS discount: ${BLUE}\$$(echo "scale=2; $TOTAL_BEFORE/100" | bc)${NC}"
echo ""

# Appliquer AUTO_COMPETE
echo -e "${YELLOW}[6/6]${NC} Test 2: Application AUTO_COMPETE..."
echo -e "${YELLOW}💡 Regardez les logs du merchant SuperShop (Terminal) !${NC}"
echo ""

sleep 1

CHECKOUT_UPDATED=$(curl -s -X PATCH http://localhost:8182/checkout/$CHECKOUT_ID \
    -H "Content-Type: application/json" \
    -d '{
        "discount_codes": ["AUTO_COMPETE"]
    }')

TOTAL_AFTER=$(echo "$CHECKOUT_UPDATED" | jq -r '.totals[] | select(.type=="total") | .amount')
DISCOUNT=$(echo "$CHECKOUT_UPDATED" | jq -r '.discounts.applied[]? | select(.code=="AUTO_COMPETE") | .amount')

if [ -z "$DISCOUNT" ] || [ "$DISCOUNT" == "null" ]; then
    echo -e "${RED}❌ AUTO_COMPETE n'a pas été appliqué${NC}"
    echo -e "${YELLOW}💡 Le code actuel n'utilise peut-être pas la nouvelle architecture${NC}"
    echo -e "${YELLOW}   Voir: sample_implementation/main_with_multiagent.go.example${NC}"
    exit 1
fi

echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║               RÉSULTATS DU TEST                  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Prix de base SuperShop    : ${BLUE}\$$(echo "scale=2; $SUPERSH0P_PRICE/100" | bc)${NC}"
echo -e "  Prix concurrent (MegaMart): ${BLUE}\$$(echo "scale=2; $MEGAMART_PRICE/100" | bc)${NC}"
echo ""
echo -e "  Total AVANT AUTO_COMPETE  : ${YELLOW}\$$(echo "scale=2; $TOTAL_BEFORE/100" | bc)${NC}"
echo -e "  Discount appliqué         : ${GREEN}-\$$(echo "scale=2; $DISCOUNT/100" | bc)${NC}"
echo -e "  Total APRÈS AUTO_COMPETE  : ${GREEN}\$$(echo "scale=2; $TOTAL_AFTER/100" | bc)${NC}"
echo ""

# Vérifier qu'on bat le concurrent
if [ $TOTAL_AFTER -lt $MEGAMART_PRICE ]; then
    echo -e "${GREEN}✅ SUCCESS: Prix final bat le concurrent !${NC}"
    SAVINGS=$((MEGAMART_PRICE - TOTAL_AFTER))
    echo -e "${GREEN}   Économie vs concurrent: \$$(echo "scale=2; $SAVINGS/100" | bc)${NC}"
else
    echo -e "${YELLOW}⚠️  Prix final ne bat pas le concurrent${NC}"
    echo -e "${YELLOW}   (Peut arriver si contraintes de marge)${NC}"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}💡 Regardez les logs du merchant SuperShop pour voir:${NC}"
echo -e "   - Agent 1 (Price Intelligence): prix concurrents"
echo -e "   - Agent 2 (Market Analysis): position marché"
echo -e "   - Agent 3 (Strategy): stratégie choisie + raisonnement"
echo -e "   - Agent 4 (Validator): validation marge"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
