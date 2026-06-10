#!/bin/bash

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  TEST EN DIRECT - SYSTÈME MULTI-AGENTS                  ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Vérifie que les services tournent
echo "🔍 Vérification des services..."
if ! curl -s http://localhost:8888 > /dev/null 2>&1; then
    echo "❌ Arena n'est pas démarré !"
    echo "Lance d'abord : ./run_unified_demo.sh"
    exit 1
fi
echo "✓ Arena tourne"

if ! curl -s http://localhost:9000/health > /dev/null 2>&1; then
    echo "❌ Shopping Graph n'est pas démarré !"
    echo "Lance d'abord : ./run_unified_demo.sh"
    exit 1
fi
echo "✓ Shopping Graph tourne"
echo ""

# Liste les marchands disponibles
echo "📋 Marchands disponibles :"
echo ""
MERCHANTS=$(curl -s http://localhost:8888/ | grep -o 'href="/[^/]*/dashboard"' | sed 's/href="\/\([^/]*\)\/dashboard"/\1/' | sort -u)

if [ -z "$MERCHANTS" ]; then
    echo "❌ Aucun marchand trouvé !"
    echo ""
    echo "Crée un marchand d'abord :"
    echo "  1. Va sur http://localhost:8888"
    echo "  2. Crée un marchand"
    echo "  3. Relance ce script"
    exit 1
fi

# Affiche les marchands
COUNT=1
declare -a MERCHANT_ARRAY
for merchant in $MERCHANTS; do
    echo "  [$COUNT] $merchant"
    MERCHANT_ARRAY[$COUNT]=$merchant
    COUNT=$((COUNT + 1))
done

echo ""
echo -n "Choisis un marchand (1-$((COUNT-1))) : "
read CHOICE

if [ -z "$CHOICE" ] || [ "$CHOICE" -lt 1 ] || [ "$CHOICE" -ge $COUNT ]; then
    echo "❌ Choix invalide"
    exit 1
fi

TENANT=${MERCHANT_ARRAY[$CHOICE]}
echo ""
echo "✓ Marchand sélectionné : $TENANT"
echo ""

# Teste l'API du marchand
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 TEST DU SYSTÈME MULTI-AGENTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Appel de l'API avec competitive pricing..."
echo ""

RESPONSE=$(curl -s -X POST "http://localhost:8888/$TENANT/api/test-auto-compete" 2>&1)

if echo "$RESPONSE" | grep -q "error\|404\|500"; then
    echo "⚠️  Réponse de l'API :"
    echo "$RESPONSE"
    echo ""
    echo "Le système multi-agents n'est peut-être pas activé."
    echo "Va sur : http://localhost:8888/$TENANT/dashboard"
    echo "Et vérifie que 'Competitive pricing' est activé"
else
    echo "✅ Réponse reçue !"
    echo ""
    echo "$RESPONSE" | jq -r '
        "👤 AGENT 2: CUSTOMER GROWTH",
        "   Analyse client en cours...",
        "",
        "📊 AGENT 3: COMPÉTITIVITÉ",
        "   Message: " + (.agent2_message // "Analyse de marché"),
        "",
        "🎯 AGENT 1: VENDEUR (DÉCISION)",
        "   " + (.main_message // "Décision finale")
    ' 2>/dev/null || echo "$RESPONSE"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Voir les logs détaillés :"
echo "   tail -f /tmp/arena.log | grep Agent"
echo ""
echo "🌐 Dashboard du marchand :"
echo "   http://localhost:8888/$TENANT/dashboard"
echo ""
