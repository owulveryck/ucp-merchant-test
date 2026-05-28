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
