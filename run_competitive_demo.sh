#!/bin/bash

# Arrêter les processus précédents s'ils existent
killall shopping-graph 2>/dev/null
killall arena 2>/dev/null

echo "🚀 Lancement de la démo avec Intelligence Compétitive"
echo ""

# Lancer Shopping Graph en arrière-plan
echo "📊 Démarrage du Shopping Graph..."
go run ./demo/cmd/shopping-graph --port 9000 --dynamic --poll-interval 10s > /tmp/shopping-graph.log 2>&1 &
SG_PID=$!

sleep 2

# Lancer Arena en premier plan
echo "🏟️  Démarrage de l'Arena..."
echo ""
echo "┌────────────────────────────────────────────────────────────────┐"
echo "│  🎯 Intelligence Compétitive ACTIVÉE                           │"
echo "│                                                                 │"
echo "│  📍 Dashboard Arena:  http://localhost:8888/                   │"
echo "│                                                                 │"
echo "│  💡 Ce qui est nouveau:                                        │"
echo "│     • Section 'Intelligence Compétitive' dans chaque dashboard │"
echo "│     • Affichage des prix des concurrents                       │"
echo "│     • Recommandations de prix en temps réel                    │"
echo "│     • Bouton 'Appliquer ce prix' pour ajuster instantanément   │"
echo "│                                                                 │"
echo "│  🎮 Comment l'utiliser:                                        │"
echo "│     1. Créez 2-3 marchands via l'interface web                 │"
echo "│     2. Configurez des prix différents                          │"
echo "│     3. Consultez la section Intelligence Compétitive           │"
echo "│     4. Cliquez 'Appliquer ce prix' pour battre vos concurrents │"
echo "│                                                                 │"
echo "│  🏆 Objectif: Avoir toujours le meilleur prix!                 │"
echo "│                                                                 │"
echo "│  ⏹️  Pour arrêter: Ctrl+C                                      │"
echo "└────────────────────────────────────────────────────────────────┘"
echo ""

go run ./demo/cmd/arena --port 8888 --graph-url http://localhost:9000 --cost-price 5000 --product-name "Casque Audio"

# Nettoyage à la sortie
echo ""
echo "🛑 Arrêt des services..."
kill $SG_PID 2>/dev/null
killall shopping-graph 2>/dev/null
echo "✅ Démo arrêtée"
