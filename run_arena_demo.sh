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

echo -e "${GREEN}✅ Tous les services sont lancés !${NC}"
echo ""
echo "┌──────────────────────────────────────────────────────────┐"
echo "│  🌐 Ouvrez dans votre navigateur:                        │"
echo "│     http://localhost:8080/                               │"
echo "│                                                           │"
echo "│  📝 ÉTAPES:                                              │"
echo "│     1. Créez 2-3 marchands                               │"
echo "│     2. Configurez des prix différents                    │"
echo "│     3. Testez AUTO_COMPETE avec un checkout              │"
echo "│                                                           │"
echo "│  🤖 Pour tester AUTO_COMPETE :                           │"
echo "│     - Dans l'interface Arena, créez un checkout          │"
echo "│     - Utilisez le code promo: AUTO_COMPETE               │"
echo "│     - Regardez le prix s'ajuster automatiquement !       │"
echo "│                                                           │"
echo "│  📊 Intelligence Compétitive:                            │"
echo "│     - Onglet \"Competitive Intel\" dans le dashboard      │"
echo "│     - Voir les prix concurrents en temps réel            │"
echo "│                                                           │"
echo "│  📝 Voir les logs des agents:                            │"
echo "│     tail -f logs/arena.log                               │"
echo "│                                                           │"
echo "│  🛑 Pour arrêter: Ctrl+C                                 │"
echo "└──────────────────────────────────────────────────────────┘"
echo ""

# Fonction pour arrêter proprement
cleanup() {
    echo ""
    echo -e "${YELLOW}🛑 Arrêt de tous les services...${NC}"

    # Arrêter les processus
    if [ -n "$SHOPPING_GRAPH_PID" ]; then
        kill $SHOPPING_GRAPH_PID 2>/dev/null || true
    fi
    if [ -n "$ARENA_PID" ]; then
        kill $ARENA_PID 2>/dev/null || true
    fi

    # Forcer si nécessaire
    pkill -f "shopping-graph" 2>/dev/null || true
    pkill -f "cmd/arena" 2>/dev/null || true

    rm -f .pids

    echo -e "${GREEN}✅ Tous les services sont arrêtés${NC}"
    echo ""
    echo -e "${BLUE}Logs sauvegardés dans logs/${NC}"
    echo "  - logs/shopping-graph.log"
    echo "  - logs/arena.log"
    echo ""
    exit 0
}

# Capturer Ctrl+C
trap cleanup SIGINT SIGTERM

# Attendre indéfiniment (les services tournent en arrière-plan)
echo -e "${YELLOW}Appuyez sur Ctrl+C pour arrêter tous les services${NC}"
echo ""
wait
