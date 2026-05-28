#!/bin/bash

# Script pour envoyer une commande d'achat au Client Agent

echo "🛒 Envoi d'une commande d'achat au Client Agent..."
echo ""

curl -X POST http://localhost:9002/command \
  -H "Content-Type: application/json" \
  -d '{"query": "I want to buy headphones", "budget": 100}'

echo ""
echo ""
echo "✅ Commande envoyée !"
echo ""
echo "👀 Regardez maintenant :"
echo "   • Les logs du Client Agent dans le terminal"
echo "   • http://localhost:9002/arena (classement en temps réel)"
echo "   • Le dashboard du marchand qui va gagner (animation VENDU !)"
echo ""
