#!/bin/bash
# Démo complète avec Agent Acheteur + Dashboard d'Observabilité

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║    Démo Complète - Agent Acheteur Multi-Agents  ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Nettoyer les anciens processus
echo -e "${YELLOW}🧹 Nettoyage des anciens processus...${NC}"
killall shopping-graph arena obs-hub client 2>/dev/null || true
pkill -f "cmd/shopping-graph" 2>/dev/null || true
pkill -f "cmd/arena" 2>/dev/null || true
pkill -f "cmd/obs-hub" 2>/dev/null || true
pkill -f "cmd/client" 2>/dev/null || true
sleep 1

# Créer le dossier logs
mkdir -p logs

echo -e "${YELLOW}🚀 Lancement de tous les services...${NC}"
echo ""

# 1. Shopping Graph
echo -e "${YELLOW}[1/4]${NC} Lancement Shopping Graph (port 9000)..."
go run ./demo/cmd/shopping-graph --port 9000 --dynamic --poll-interval 10s > logs/shopping-graph.log 2>&1 &
SHOPPING_PID=$!
sleep 2

if ! ps -p $SHOPPING_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Shopping Graph n'a pas démarré${NC}"
    cat logs/shopping-graph.log
    exit 1
fi
echo -e "${GREEN}✓ Shopping Graph démarré (PID: $SHOPPING_PID)${NC}"

# 2. Observability Hub
echo -e "${YELLOW}[2/4]${NC} Lancement Observability Hub (port 9002)..."
go run ./demo/cmd/obs-hub --port 9002 --graph-url http://localhost:9000 --arena-url http://localhost:8888 > logs/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2

if ! ps -p $OBS_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Obs Hub n'a pas démarré${NC}"
    cat logs/obs-hub.log
    exit 1
fi
echo -e "${GREEN}✓ Obs Hub démarré (PID: $OBS_PID)${NC}"

# 3. Arena
echo -e "${YELLOW}[3/4]${NC} Lancement Arena (port 8888)..."
go run ./demo/cmd/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 > logs/arena.log 2>&1 &
ARENA_PID=$!
sleep 3

if ! ps -p $ARENA_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Arena n'a pas démarré${NC}"
    cat logs/arena.log
    exit 1
fi
echo -e "${GREEN}✓ Arena démarré (PID: $ARENA_PID)${NC}"

# 4. Client Agent
echo -e "${YELLOW}[4/4]${NC} Lancement Client Agent (Gemini)..."
go run ./demo/cmd/client --obs-url http://localhost:9002 > logs/client.log 2>&1 &
CLIENT_PID=$!
sleep 2

if ! ps -p $CLIENT_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Client Agent n'a pas démarré${NC}"
    cat logs/client.log
    exit 1
fi
echo -e "${GREEN}✓ Client Agent démarré (PID: $CLIENT_PID)${NC}"

# Sauvegarder les PIDs
echo "$SHOPPING_PID" > .pids
echo "$OBS_PID" >> .pids
echo "$ARENA_PID" >> .pids
echo "$CLIENT_PID" >> .pids

echo ""
echo -e "${GREEN}✅ Tous les services sont lancés !${NC}"
echo ""
echo "┌──────────────────────────────────────────────────────────┐"
echo "│  🌐 Ouvrez dans votre navigateur:                        │"
echo "│     http://localhost:8888/                               │"
echo "│                                                           │"
echo "│  📝 ÉTAPES:                                              │"
echo "│     1. Créez 2-3 marchands                               │"
echo "│     2. Configurez des prix différents                    │"
echo "│     3. Revenez ici et tapez: ./acheter.sh                │"
echo "│                                                           │"
echo "│  👀 Dashboard d'Observabilité (voir l'agent penser):     │"
echo "│     http://localhost:9002/arena                          │"
echo "│                                                           │"
echo "│  🤖 L'agent acheteur va:                                 │"
echo "│     - Chercher dans le Shopping Graph                    │"
echo "│     - Comparer les prix des marchands                    │"
echo "│     - Créer des checkouts avec AUTO_COMPETE              │"
echo "│     - Acheter chez le moins cher                         │"
echo "│                                                           │"
echo "│  🛑 Pour arrêter: Ctrl+C                                 │"
echo "└──────────────────────────────────────────────────────────┘"
echo ""

# Créer le script acheter.sh
cat > acheter.sh << 'EOFSCRIPT'
#!/bin/bash
echo "🛒 Envoi de la commande d'achat à l'agent..."
echo ""
curl -s -X POST http://localhost:9002/command \
  -H "Content-Type: application/json" \
  -d '{"query": "I want to buy headphones", "budget": 100}'
echo ""
echo ""
echo "✅ Commande envoyée !"
echo ""
echo "👀 Regardez le dashboard pour voir l'agent penser:"
echo "   http://localhost:9002/arena"
echo ""
echo "📝 Ou les logs en temps réel:"
echo "   tail -f logs/client.log"
EOFSCRIPT
chmod +x acheter.sh

# Fonction pour arrêter proprement
cleanup() {
    echo ""
    echo -e "${YELLOW}🛑 Arrêt de tous les services...${NC}"
    kill $SHOPPING_PID $OBS_PID $ARENA_PID $CLIENT_PID 2>/dev/null || true
    killall shopping-graph arena obs-hub client 2>/dev/null || true
    rm -f .pids
    echo -e "${GREEN}✅ Tous les services sont arrêtés${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${YELLOW}Appuyez sur Ctrl+C pour arrêter tous les services${NC}"
echo ""
wait
