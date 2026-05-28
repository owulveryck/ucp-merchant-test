#!/bin/bash
# Arrêt de tous les services de la démo

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}🛑 Arrêt de tous les services...${NC}"
echo ""

# Lire les PIDs sauvegardés
if [ -f .pids ]; then
    while read pid; do
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${YELLOW}Arrêt du processus $pid...${NC}"
            kill $pid 2>/dev/null || true
        fi
    done < .pids
    rm .pids
fi

# Forcer l'arrêt si nécessaire
pkill -f "shopping-graph" 2>/dev/null || true
pkill -f "sample_implementation.*8182" 2>/dev/null || true
pkill -f "sample_implementation.*8183" 2>/dev/null || true
pkill -f "sample_implementation.*8184" 2>/dev/null || true
pkill -f "cmd/arena" 2>/dev/null || true

sleep 1

echo -e "${GREEN}✅ Tous les services sont arrêtés${NC}"
echo ""

# Vérifier qu'ils sont bien arrêtés
if lsof -i :9000 > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Port 9000 encore occupé${NC}"
fi
if lsof -i :8182 > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Port 8182 encore occupé${NC}"
fi
if lsof -i :8183 > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Port 8183 encore occupé${NC}"
fi
if lsof -i :8184 > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Port 8184 encore occupé${NC}"
fi

echo -e "${GREEN}Logs sauvegardés dans logs/${NC}"
echo -e "  - logs/shopping-graph.log"
echo -e "  - logs/superShop.log"
echo -e "  - logs/megaMart.log"
echo -e "  - logs/budgetBuy.log"
