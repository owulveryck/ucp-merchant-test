#!/bin/bash
# Lancement automatique de la démo multi-agents

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Démarrage Démo Multi-Agents AUTO_COMPETE    ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Nettoyer les anciens processus
echo -e "${YELLOW}🧹 Nettoyage des anciens processus...${NC}"
pkill -f "shopping-graph" 2>/dev/null || true
pkill -f "sample_implementation.*8182" 2>/dev/null || true
pkill -f "sample_implementation.*8183" 2>/dev/null || true
pkill -f "sample_implementation.*8184" 2>/dev/null || true
sleep 1

# Créer le dossier logs
mkdir -p logs

# Lancer Shopping Graph
echo -e "${YELLOW}🚀 Lancement Shopping Graph (port 9000)...${NC}"
cd demo
go run ./cmd/shopping-graph --port 9000 > ../logs/shopping-graph.log 2>&1 &
SHOPPING_GRAPH_PID=$!
cd ..
sleep 2

# Vérifier que le Shopping Graph tourne
if ! ps -p $SHOPPING_GRAPH_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: Shopping Graph n'a pas démarré${NC}"
    cat logs/shopping-graph.log
    exit 1
fi
echo -e "${GREEN}✓ Shopping Graph démarré (PID: $SHOPPING_GRAPH_PID)${NC}"

# Lancer SuperShop
echo -e "${YELLOW}🚀 Lancement SuperShop (port 8182)...${NC}"
go run ./sample_implementation \
    --port 8182 \
    --data-dir demo/data/merchant_a \
    --data-format json \
    --merchant-name SuperShop \
    > logs/superShop.log 2>&1 &
SUPERSH0P_PID=$!
sleep 2

if ! ps -p $SUPERSH0P_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: SuperShop n'a pas démarré${NC}"
    cat logs/superShop.log
    exit 1
fi
echo -e "${GREEN}✓ SuperShop démarré (PID: $SUPERSH0P_PID)${NC}"

# Lancer MegaMart
echo -e "${YELLOW}🚀 Lancement MegaMart (port 8183)...${NC}"
go run ./sample_implementation \
    --port 8183 \
    --data-dir demo/data/merchant_b \
    --data-format json \
    --merchant-name MegaMart \
    > logs/megaMart.log 2>&1 &
MEGAMART_PID=$!
sleep 2

if ! ps -p $MEGAMART_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: MegaMart n'a pas démarré${NC}"
    cat logs/megaMart.log
    exit 1
fi
echo -e "${GREEN}✓ MegaMart démarré (PID: $MEGAMART_PID)${NC}"

# Lancer BudgetBuy
echo -e "${YELLOW}🚀 Lancement BudgetBuy (port 8184)...${NC}"
go run ./sample_implementation \
    --port 8184 \
    --data-dir demo/data/merchant_c \
    --data-format json \
    --merchant-name BudgetBuy \
    > logs/budgetBuy.log 2>&1 &
BUDGETBUY_PID=$!
sleep 2

if ! ps -p $BUDGETBUY_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: BudgetBuy n'a pas démarré${NC}"
    cat logs/budgetBuy.log
    exit 1
fi
echo -e "${GREEN}✓ BudgetBuy démarré (PID: $BUDGETBUY_PID)${NC}"

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║            Tous les services sont lancés         ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Shopping Graph: http://localhost:9000  (PID: $SHOPPING_GRAPH_PID)"
echo -e "  SuperShop:      http://localhost:8182  (PID: $SUPERSH0P_PID)"
echo -e "  MegaMart:       http://localhost:8183  (PID: $MEGAMART_PID)"
echo -e "  BudgetBuy:      http://localhost:8184  (PID: $BUDGETBUY_PID)"
echo ""
echo -e "${YELLOW}⏳ Attente de 30s pour l'indexation...${NC}"

# Barre de progression
for i in {1..30}; do
    printf "${BLUE}█${NC}"
    sleep 1
done
echo ""
echo ""

echo -e "${GREEN}✅ Indexation terminée !${NC}"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Que veux-tu faire ?${NC}"
echo ""
echo -e "  1️⃣  ${GREEN}Lancer le test automatique${NC}"
echo -e "     ${YELLOW}→ ./test_multi_agent.sh${NC}"
echo ""
echo -e "  2️⃣  ${GREEN}Test manuel avec curl${NC}"
echo -e "     ${YELLOW}→ Voir les commandes ci-dessous${NC}"
echo ""
echo -e "  3️⃣  ${GREEN}Voir les logs en temps réel${NC}"
echo -e "     ${YELLOW}→ tail -f logs/superShop.log${NC}"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Sauvegarder les PIDs
echo "$SHOPPING_GRAPH_PID" > .pids
echo "$SUPERSH0P_PID" >> .pids
echo "$MEGAMART_PID" >> .pids
echo "$BUDGETBUY_PID" >> .pids

echo -e "${YELLOW}💡 Test manuel rapide :${NC}"
echo ""
echo -e "${BLUE}# 1. Créer un checkout${NC}"
echo -e 'CHECKOUT=$(curl -s -X POST http://localhost:8182/checkout -H "Content-Type: application/json" -d '\''{"items":[{"product_id":"prod_roses_bouquet","quantity":1}]}'\'' | jq -r ".id")'
echo ""
echo -e "${BLUE}# 2. Appliquer AUTO_COMPETE${NC}"
echo -e 'curl -X PATCH http://localhost:8182/checkout/$CHECKOUT -H "Content-Type: application/json" -d '\''{"discount_codes":["AUTO_COMPETE"]}'\'' | jq'
echo ""
echo -e "${BLUE}# 3. Voir les logs des agents${NC}"
echo -e 'tail -f logs/superShop.log'
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}Pour arrêter tous les services :${NC}"
echo -e "  ${YELLOW}./stop_demo.sh${NC}"
echo ""
