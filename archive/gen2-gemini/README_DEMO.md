# 🚀 Démo Multi-Agents en 2 Commandes

## ⚡ Lancer la Démo

### Une seule ligne !

```bash
./run_multiagent_demo.sh
```

Ce script va :
- ✅ Lancer Shopping Graph (port 9000)
- ✅ Lancer SuperShop (port 8182)
- ✅ Lancer MegaMart (port 8183)
- ✅ Lancer BudgetBuy (port 8184)
- ✅ Attendre 30s pour l'indexation
- ✅ Te donner les instructions pour tester

**Durée** : ~40 secondes

---

## 🧪 Tester AUTO_COMPETE

### Une seule ligne !

```bash
./test_multi_agent.sh
```

Ce script va :
- ✅ Vérifier que tout tourne
- ✅ Créer un checkout
- ✅ Appliquer AUTO_COMPETE
- ✅ Afficher les résultats

**Résultat attendu** :
```
Prix concurrent (MegaMart): $59.99
Total AVANT AUTO_COMPETE  : $65.00
Discount appliqué         : -$9.24
Total APRÈS AUTO_COMPETE  : $55.76

✅ SUCCESS: Prix final bat le concurrent !
```

---

## 🔍 Voir les Logs des Agents

```bash
tail -f logs/superShop.log
```

Tu verras les **4 agents** en action :

```
[Orchestrator] Starting competitive pricing analysis...
[Orchestrator] Price Intelligence: rank 2/3, lowest: $59.99 (MegaMart)
[Orchestrator] Market Analysis: follower position, stable trend
[Orchestrator] Strategy: balanced, target: $56.99, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]
[Orchestrator] ✅ Pricing approved: $56.99 (discount: $9.24, margin: 25%)
```

---

## 🛑 Arrêter la Démo

```bash
./stop_demo.sh
```

Arrête proprement tous les services.

---

## 📊 Test Manuel (optionnel)

Si tu veux tester manuellement sans le script :

```bash
# 1. Créer un checkout
CHECKOUT=$(curl -s -X POST http://localhost:8182/checkout \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":"prod_roses_bouquet","quantity":1}]}' \
  | jq -r '.id')

echo "Checkout: $CHECKOUT"

# 2. Voir le prix AVANT
curl http://localhost:8182/checkout/$CHECKOUT | jq '.totals'

# 3. Appliquer AUTO_COMPETE
curl -X PATCH http://localhost:8182/checkout/$CHECKOUT \
  -H "Content-Type: application/json" \
  -d '{"discount_codes":["AUTO_COMPETE"]}' \
  | jq '.totals'

# 4. Regarder les logs
tail -f logs/superShop.log
```

---

## 🗂️ Fichiers de Logs

Tous les logs sont dans `logs/` :

```
logs/
├── shopping-graph.log   # Shopping Graph
├── superShop.log        # SuperShop (REGARDER ICI pour les agents!)
├── megaMart.log         # MegaMart
└── budgetBuy.log        # BudgetBuy
```

---

## ❓ Problèmes ?

### "Port already in use"

```bash
./stop_demo.sh
./run_multiagent_demo.sh
```

### "Shopping Graph not indexing"

Attends 30 secondes de plus. L'indexation est toutes les 30s.

### "AUTO_COMPETE not applied"

Le code actuel n'utilise pas encore la nouvelle architecture multi-agents.
Voir : `sample_implementation/main_with_multiagent.go.example`

---

## 📚 Documentation Complète

- `QUICKSTART_MULTIAGENT.md` - Guide rapide
- `LAUNCH_MULTI_AGENT.md` - Guide détaillé
- `pkg/merchant/competitive/INTEGRATION.md` - Intégration dans le code
- `docs/adr/003-competitive-pricing-agent.md` - Architecture Decision Record

---

## 🎯 En Résumé

```bash
# Lancer tout
./run_multiagent_demo.sh

# Tester AUTO_COMPETE
./test_multi_agent.sh

# Voir les agents en action
tail -f logs/superShop.log

# Arrêter tout
./stop_demo.sh
```

**C'est tout !** 🎉
