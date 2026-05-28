#!/bin/bash
# Lancement de l'Arena mode avec contrôle des prix

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Démarrage Arena Mode - Multi-Agents     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Nettoyer les anciens processus
echo -e "${YELLOW}🧹 Nettoyage des anciens processus...${NC}"
pkill -f "shopping-graph" 2>/dev/null || true
pkill -f "arena" 2>/dev/null || true
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

# Lancer Arena
echo -e "${YELLOW}🚀 Lancement Arena (port 8080)...${NC}"
cd demo
go run ./cmd/arena --port 8080 > ../logs/arena.log 2>&1 &
ARENA_PID=$!
cd ..
sleep 3

if ! ps -p $ARENA_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: Arena n'a pas démarré${NC}"
    cat logs/arena.log
    exit 1
fi
echo -e "${GREEN}✓ Arena démarré (PID: $ARENA_PID)${NC}"

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              Arena Mode Lancé !                  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Shopping Graph: ${BLUE}http://localhost:9000${NC}  (PID: $SHOPPING_GRAPH_PID)"
echo -e "  Arena:          ${BLUE}http://localhost:8080${NC}  (PID: $ARENA_PID)"
echo ""
echo -e "${YELLOW}⏳ Attente de 5s pour initialisation...${NC}"

# Barre de progression
for i in {1..5}; do
    printf "${BLUE}█${NC}"
    sleep 1
done
echo ""
echo ""

# Sauvegarder les PIDs
echo "$SHOPPING_GRAPH_PID" > .pids
echo "$ARENA_PID" >> .pids

echo -e "${GREEN}✅ Tout est prêt !${NC}"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}📊 Interface Arena :${NC}"
echo -e ""
echo -e "   Ouvre ton navigateur sur : ${GREEN}http://localhost:8080${NC}"
echo -e ""
echo -e "   Tu peux :"
echo -e "   ✅ Créer des merchants (SuperShop, MegaMart, etc.)"
echo -e "   ✅ Ajuster les prix en temps réel"
echo -e "   ✅ Modifier les codes promo"
echo -e "   ✅ Changer les options de livraison"
echo -e "   ✅ Voir l'activité en direct"
echo -e ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}🤖 Pour tester AUTO_COMPETE :${NC}"
echo ""
echo -e "   1. Crée 2-3 merchants dans l'interface Arena"
echo -e "   2. Définis leurs prix (ex: MegaMart moins cher)"
echo -e "   3. Crée un checkout avec code ${GREEN}AUTO_COMPETE${NC}"
echo -e "   4. Regarde les logs : ${YELLOW}tail -f logs/arena.log${NC}"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}Pour arrêter :${NC}"
echo -e "  ${YELLOW}./stop_demo.sh${NC}"
echo ""
