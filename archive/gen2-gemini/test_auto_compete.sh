#!/bin/bash
# Test rapide du code AUTO_COMPETE

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Test AUTO_COMPETE - Multi-Agents        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Vérifier que Arena tourne
if ! curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${RED}❌ Arena n'est pas accessible sur :8080${NC}"
    echo -e "${YELLOW}💡 Lancez d'abord : ./run_arena_demo.sh${NC}"
    exit 1
fi

# Lister les tenants (merchants)
echo -e "${YELLOW}[1/4]${NC} Récupération de la liste des merchants..."
TENANTS=$(curl -s http://localhost:8080/api/tenants | jq -r '.[].id')

if [ -z "$TENANTS" ]; then
    echo -e "${RED}❌ Aucun merchant trouvé${NC}"
    echo -e "${YELLOW}💡 Créez des merchants dans l'interface Arena (http://localhost:8080)${NC}"
    exit 1
fi

TENANT_COUNT=$(echo "$TENANTS" | wc -l | tr -d ' ')
echo -e "${GREEN}✓ $TENANT_COUNT merchant(s) trouvé(s)${NC}"

# Prendre le premier tenant
TENANT_ID=$(echo "$TENANTS" | head -1)
TENANT_NAME=$(curl -s http://localhost:8080/api/tenants | jq -r ".[] | select(.id==\"$TENANT_ID\") | .name")

echo -e "${BLUE}  → Utilisation de: $TENANT_NAME (ID: $TENANT_ID)${NC}"
echo ""

# Lister les produits
echo -e "${YELLOW}[2/4]${NC} Récupération des produits..."
PRODUCTS=$(curl -s http://localhost:8080/$TENANT_ID/api/products)
PRODUCT_COUNT=$(echo "$PRODUCTS" | jq '. | length')

if [ "$PRODUCT_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ Aucun produit trouvé chez $TENANT_NAME${NC}"
    exit 1
fi

echo -e "${GREEN}✓ $PRODUCT_COUNT produit(s) disponible(s)${NC}"

# Prendre le premier produit
PRODUCT_ID=$(echo "$PRODUCTS" | jq -r '.[0].id')
PRODUCT_NAME=$(echo "$PRODUCTS" | jq -r '.[0].name')
PRODUCT_PRICE=$(echo "$PRODUCTS" | jq -r '.[0].price')

echo -e "${BLUE}  → Produit: $PRODUCT_NAME${NC}"
echo -e "${BLUE}  → Prix de base: \$$(echo "scale=2; $PRODUCT_PRICE/100" | bc)${NC}"
echo ""

# Créer un checkout SANS AUTO_COMPETE
echo -e "${YELLOW}[3/4]${NC} Création d'un checkout SANS AUTO_COMPETE..."

CHECKOUT_RESPONSE=$(curl -s -X POST http://localhost:8080/$TENANT_ID/checkout \
    -H "Content-Type: application/json" \
    -d "{
        \"items\": [{
            \"product_id\": \"$PRODUCT_ID\",
            \"quantity\": 1
        }]
    }")

CHECKOUT_ID=$(echo "$CHECKOUT_RESPONSE" | jq -r '.id')
TOTAL_BEFORE=$(echo "$CHECKOUT_RESPONSE" | jq -r '.totals[] | select(.type=="total") | .amount')

echo -e "${GREEN}✓ Checkout créé: $CHECKOUT_ID${NC}"
echo -e "${BLUE}  → Total SANS discount: \$$(echo "scale=2; $TOTAL_BEFORE/100" | bc)${NC}"
echo ""

# Appliquer AUTO_COMPETE
echo -e "${YELLOW}[4/4]${NC} Application du code AUTO_COMPETE..."
echo -e "${YELLOW}💡 Regardez les logs: tail -f logs/arena.log${NC}"
echo ""

sleep 1

CHECKOUT_UPDATED=$(curl -s -X PATCH http://localhost:8080/$TENANT_ID/checkout/$CHECKOUT_ID \
    -H "Content-Type: application/json" \
    -d '{
        "discount_codes": ["AUTO_COMPETE"]
    }')

TOTAL_AFTER=$(echo "$CHECKOUT_UPDATED" | jq -r '.totals[] | select(.type=="total") | .amount')
DISCOUNT=$(echo "$CHECKOUT_UPDATED" | jq -r '.discounts.applied[]? | select(.code=="AUTO_COMPETE") | .amount')

if [ -z "$DISCOUNT" ] || [ "$DISCOUNT" == "null" ]; then
    echo -e "${YELLOW}⚠️  AUTO_COMPETE n'a pas appliqué de discount${NC}"
    echo ""
    echo -e "${BLUE}Raisons possibles :${NC}"
    echo -e "  - Vous êtes déjà le moins cher"
    echo -e "  - Pas assez de merchants concurrents"
    echo -e "  - Contraintes de marge"
    echo ""
    echo -e "${YELLOW}💡 Créez plus de merchants avec des prix différents dans l'Arena${NC}"
    exit 0
fi

echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║               RÉSULTATS DU TEST                  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Merchant              : ${BLUE}$TENANT_NAME${NC}"
echo -e "  Produit               : ${BLUE}$PRODUCT_NAME${NC}"
echo ""
echo -e "  Total AVANT AUTO_COMPETE  : ${YELLOW}\$$(echo "scale=2; $TOTAL_BEFORE/100" | bc)${NC}"
echo -e "  Discount appliqué         : ${GREEN}-\$$(echo "scale=2; $DISCOUNT/100" | bc)${NC}"
echo -e "  Total APRÈS AUTO_COMPETE  : ${GREEN}\$$(echo "scale=2; $TOTAL_AFTER/100" | bc)${NC}"
echo ""

SAVINGS=$((TOTAL_BEFORE - TOTAL_AFTER))
echo -e "  ${GREEN}✅ Économie: \$$(echo "scale=2; $SAVINGS/100" | bc)${NC}"
echo ""

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}💡 Consultez les logs pour voir les 4 agents :${NC}"
echo -e "   tail -f logs/arena.log"
echo ""
echo -e "${YELLOW}💡 Testez avec différents merchants et prix dans l'Arena !${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
