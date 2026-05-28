#!/bin/bash

# Arrêter les anciens processus
killall shopping-graph arena obs-hub client 2>/dev/null

echo "🚀 Démarrage de tous les services..."
echo ""

# Lancer tous les services en arrière-plan
go run ./demo/cmd/shopping-graph --port 9000 --dynamic --poll-interval 10s > /dev/null 2>&1 &
sleep 2

go run ./demo/cmd/obs-hub --port 9002 --graph-url http://localhost:9000 --arena-url http://localhost:8888 > /dev/null 2>&1 &
sleep 2

go run ./demo/cmd/arena --port 8888 --graph-url http://localhost:9000 --obs-url http://localhost:9002 --cost-price 5000 > /dev/null 2>&1 &
sleep 3

go run ./demo/cmd/client --obs-url http://localhost:9002 > /dev/null 2>&1 &
sleep 2

echo "✅ Tous les services sont lancés !"
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
echo "│  👀 Voir le classement:                                  │"
echo "│     http://localhost:9002/arena                          │"
echo "│                                                           │"
echo "│  🛑 Pour arrêter: Ctrl+C puis: killall shopping-graph    │"
echo "└──────────────────────────────────────────────────────────┘"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter tous les services"
echo ""

# Créer le script acheter.sh
cat > acheter.sh << 'EOF'
#!/bin/bash
echo "🛒 Envoi de la commande d'achat..."
curl -s -X POST http://localhost:9002/command \
  -H "Content-Type: application/json" \
  -d '{"query": "I want to buy headphones", "budget": 100}'
echo ""
echo "✅ Commande envoyée ! Regardez http://localhost:9002/arena"
EOF
chmod +x acheter.sh

# Attendre Ctrl+C
trap "echo ''; echo '🛑 Arrêt...'; killall shopping-graph arena obs-hub client 2>/dev/null; echo '✅ Terminé'; exit" INT
wait
