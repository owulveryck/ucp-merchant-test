#!/bin/bash

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  DÉMONSTRATION SYSTÈME MULTI-AGENTS                     ║"
echo "║  4 Agents : Intel → Analysis → Strategy → Validation   ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

TENANT_ID="75cfb489"
TENANT_NAME="Test"

echo "🎯 Marchand : $TENANT_NAME"
echo "🌐 Dashboard : http://localhost:8888/$TENANT_ID/dashboard"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 APPEL DU SYSTÈME MULTI-AGENTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

RESPONSE=$(curl -s -X POST "http://localhost:8888/$TENANT_ID/api/test-auto-compete")

echo "$RESPONSE" | jq -r '
"📊 AGENT 1: PRICE INTELLIGENCE",
"   " + .reasoning.agent1,
"",
"📈 AGENT 2: MARKET ANALYSIS",
"   " + .reasoning.agent2,
"",
"🎯 AGENT 3: STRATEGY RECOMMENDER",
"   " + (.reasoning.agent3 | gsub("<br>"; "\n   ") | gsub("<strong>"; "") | gsub("</strong>"; "")),
"",
"✅ AGENT 4: MARGIN VALIDATOR",
"   " + .reasoning.agent4,
"",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"💰 DÉCISION FINALE",
"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
"",
"   Prix initial :    $" + ((.current_price / 100) | tostring),
"   Réduction :       -$" + ((.discount_amount / 100) | tostring),
"   Prix final :      $" + ((.final_price / 100) | tostring) + " ✨",
"   Marge :           " + (.margin_percent | tostring) + "%",
""
' 2>/dev/null

echo ""
echo "🎮 Pour voir en temps réel :"
echo "   1. Va sur http://localhost:8888/$TENANT_ID/dashboard"
echo "   2. Bouge le slider de prix"
echo "   3. Les agents recalculent automatiquement !"
echo ""
echo "📜 Voir les logs détaillés :"
echo "   tail -f /tmp/arena.log | grep -E 'Agent|Orchestrator'"
echo ""
