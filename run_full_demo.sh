#!/bin/bash

# Arrêter les processus précédents
killall shopping-graph arena obs-hub client 2>/dev/null

echo "🎮 Lancement de la DÉMO COMPLÈTE avec Client Agent"
echo ""

# Lancer Shopping Graph
echo "📊 Shopping Graph (port 9000)..."
go run ./demo/cmd/shopping-graph --port 9000 --dynamic --poll-interval 10s > /tmp/shopping-graph.log 2>&1 &
SG_PID=$!
sleep 2

# Lancer Obs Hub
echo "👁️  Observability Hub (port 9002)..."
go run ./demo/cmd/obs-hub --port 9002 --graph-url http://localhost:9000 --arena-url http://localhost:8888 > /tmp/obs-hub.log 2>&1 &
OBS_PID=$!
sleep 2

# Lancer Arena
echo "🏟️  Arena (port 8888)..."
go run ./demo/cmd/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 --product-name "Casque Audio" > /tmp/arena.log 2>&1 &
ARENA_PID=$!
sleep 3

# Lancer Client Agent
echo "🤖 Client Agent..."
go run ./demo/cmd/client --obs-url http://localhost:9002 > /tmp/client.log 2>&1 &
CLIENT_PID=$!
sleep 2

echo ""
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│  ✅ TOUS LES SERVICES SONT LANCÉS !                          │"
echo "│                                                               │"
echo "│  ÉTAPE 1 - Créer les marchands:                              │"
echo "│  http://localhost:8888/                                      │"
echo "│  → Cliquer 'Register' 3 fois                                 │"
echo "│  → Configurer les prix dans chaque dashboard                 │"
echo "│                                                               │"
echo "│  ÉTAPE 2 - Lancer l'agent acheteur:                          │"
echo "│  http://localhost:9002/arena                                 │"
echo "│  → Dans 'Agent Acheteur', taper votre demande:               │"
echo "│     💰 Moins cher: \"Achète un casque\"                        │"
echo "│     ⚡ Plus rapide: \"Achète un casque rapidement\"            │"
echo "│  → Cliquer 'Send' et observer !                              │"
echo "│                                                               │"
echo "│  📊 Dashboards disponibles:                                  │"
echo "│     • Arena:     http://localhost:8888/                      │"
echo "│     • Monitor:   http://localhost:9002/arena                 │"
echo "│     • Arena 2:   http://localhost:9002/arena2                │"
echo "│     • Insights:  http://localhost:9002/insights              │"
echo "│                                                               │"
echo "│  ⏹️  Pour arrêter: Ctrl+C                                     │"
echo "└──────────────────────────────────────────────────────────────┘"
echo ""
echo "🔄 Affichage des logs du Client Agent en temps réel..."
echo "   (Envoyez une commande pour voir le Client Agent en action)"
echo ""

# Afficher les logs du client agent
tail -f /tmp/client.log &
TAIL_PID=$!

# Attendre Ctrl+C
trap "echo ''; echo '🛑 Arrêt des services...'; kill $SG_PID $OBS_PID $ARENA_PID $CLIENT_PID $TAIL_PID 2>/dev/null; killall shopping-graph arena obs-hub client 2>/dev/null; echo '✅ Tous les services arrêtés'; exit" INT

wait
