#!/bin/bash
# Démo complète avec système 3-agent + Dashboard d'Observabilité

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║    Démo Système 3-Agent + Observabilité         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Nettoyer les anciens processus
echo -e "${YELLOW}🧹 Nettoyage des anciens processus...${NC}"
killall shopping-graph arena obs-hub 2>/dev/null || true
pkill -f "cmd/shopping-graph" 2>/dev/null || true
pkill -f "cmd/arena" 2>/dev/null || true
pkill -f "cmd/obs-hub" 2>/dev/null || true
sleep 1

# Créer le dossier logs et bin
mkdir -p logs bin

echo -e "${YELLOW}🔨 Compilation des binaires...${NC}"
go build -o bin/arena ./demo/cmd/arena
echo ""

echo -e "${YELLOW}🚀 Lancement de tous les services...${NC}"
echo ""

# 1. Shopping Graph
echo -e "${YELLOW}[1/3]${NC} Lancement Shopping Graph (port 9000)..."
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
echo -e "${YELLOW}[2/3]${NC} Lancement Observability Hub (port 9002)..."
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
echo -e "${YELLOW}[3/3]${NC} Lancement Arena (port 8888)..."
./bin/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 --competitive-pricing --min-margin 10 > logs/arena.log 2>&1 &
ARENA_PID=$!
sleep 3

if ! ps -p $ARENA_PID > /dev/null 2>&1; then
    echo -e "${RED}❌ Arena n'a pas démarré${NC}"
    cat logs/arena.log
    exit 1
fi
echo -e "${GREEN}✓ Arena démarré (PID: $ARENA_PID)${NC}"

# Sauvegarder les PIDs
echo "$SHOPPING_PID" > .pids
echo "$OBS_PID" >> .pids
echo "$ARENA_PID" >> .pids

# Initialiser TAIL_PID pour cleanup
TAIL_PID=""

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
echo "│     3. Testez AUTO_COMPETE dans le dashboard             │"
echo "│                                                           │"
echo "│  👀 Dashboard d'Observabilité:                           │"
echo "│     http://localhost:9002/arena                          │"
echo "│                                                           │"
echo "│  🤖 Système 3-Agent:                                     │"
echo "│     - Agent 1: Intelligence Marché                       │"
echo "│     - Agent 2: Fidélisation Client (VIP)                 │"
echo "│     - Agent 3: Décision Finale                           │"
echo "│                                                           │"
echo "│  🛑 Pour arrêter: Ctrl+C                                 │"
echo "└──────────────────────────────────────────────────────────┘"
echo ""

# Créer le script acheter.sh
cat > acheter.sh << 'EOFSCRIPT'
#!/bin/bash
echo "🧪 Test du système 3-agent sur tous les marchands..."
echo ""

# Récupère la liste des marchands
MERCHANTS=$(curl -s http://localhost:8888/merchants | jq -r '.merchants[].id')

for ID in $MERCHANTS; do
  NAME=$(curl -s http://localhost:8888/merchants | jq -r ".merchants[] | select(.id==\"$ID\") | .name")
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📊 Test pour: $NAME (ID: $ID)"
  echo ""
  curl -s -X POST http://localhost:8888/$ID/api/test-auto-compete | jq '.'
  echo ""
done

echo "✅ Test terminé !"
echo ""
echo "👀 Regardez le dashboard pour voir les agents:"
echo "   http://localhost:9002/arena"
EOFSCRIPT
chmod +x acheter.sh

# Fonction pour arrêter proprement
cleanup() {
    echo ""
    echo -e "${YELLOW}🛑 Arrêt de tous les services...${NC}"
    kill $SHOPPING_PID $OBS_PID $ARENA_PID $TAIL_PID 2>/dev/null || true
    killall shopping-graph arena obs-hub tail 2>/dev/null || true
    rm -f .pids
    echo -e "${GREEN}✅ Tous les services sont arrêtés${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${YELLOW}Appuyez sur Ctrl+C pour arrêter tous les services${NC}"
echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   📊 LOGS DES 3 AGENTS EN TEMPS RÉEL${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Attendre que le fichier de log existe
sleep 2

# Suivre les logs en temps réel, filtrer pour les 3 agents
tail -f logs/arena.log 2>/dev/null | grep --line-buffered -E "Agent|ArenaAdapter|3-AGENT|Market Intelligence|Customer Retention|Final Decision" &
TAIL_PID=$!

# Attendre
wait
