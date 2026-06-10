# 🚀 Démarrage Rapide - Multi-Agent Pricing

## ⚡ 3 Étapes Pour Tester

### 1️⃣ Lancer les Services (4 terminaux)

**Terminal 1 - Shopping Graph**
```bash
cd demo
go run ./cmd/shopping-graph --port 9000
```

**Terminal 2 - SuperShop (notre merchant)**
```bash
go run ./sample_implementation \
  --port 8182 \
  --data-dir demo/data/merchant_a \
  --merchant-name SuperShop
```

**Terminal 3 - MegaMart (concurrent 1)**
```bash
go run ./sample_implementation \
  --port 8183 \
  --data-dir demo/data/merchant_b \
  --merchant-name MegaMart
```

**Terminal 4 - BudgetBuy (concurrent 2)**
```bash
go run ./sample_implementation \
  --port 8184 \
  --data-dir demo/data/merchant_c \
  --merchant-name BudgetBuy
```

⏱️ **Attendez 30 secondes** que le Shopping Graph indexe les merchants.

---

### 2️⃣ Lancer le Test Automatique

**Terminal 5**
```bash
./test_multi_agent.sh
```

Le script va :
1. ✅ Vérifier que tous les services tournent
2. 📊 Récupérer les prix concurrents
3. 🛒 Créer un checkout SANS AUTO_COMPETE
4. 🤖 Appliquer AUTO_COMPETE
5. 📈 Afficher les résultats

---

### 3️⃣ Regarder les Logs du Terminal 2 (SuperShop)

Vous verrez les **4 agents** en action :

```
[Orchestrator] Starting competitive pricing analysis for product prod_roses_bouquet

[Agent 1 - Price Intelligence]
[Orchestrator] Price Intelligence: rank 2/3, lowest: $59.99 (MegaMart)

[Agent 2 - Market Analysis]
[Orchestrator] Market Analysis: follower position, stable trend, opportunity: optimize

[Agent 3 - Strategy Recommender]
[Orchestrator] Strategy: balanced, target: $56.99, discount: $9.24, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]

[Agent 4 - Margin Validator]
[Orchestrator] ✅ Pricing approved: $56.99 (discount: $9.24, margin: 25%)
```

---

## 🎯 Résultat Attendu

```
Prix concurrent (MegaMart): $59.99
Total AVANT AUTO_COMPETE  : $65.00
Discount appliqué         : -$9.24
Total APRÈS AUTO_COMPETE  : $55.76

✅ SUCCESS: Prix final bat le concurrent !
   Économie vs concurrent: $4.23
```

---

## 🧪 Tester d'Autres Scénarios

### Scénario 1 : Stock Bas → Stratégie Aggressive

Modifier `demo/data/merchant_a/products.csv` :
```csv
prod_roses_bouquet,Bouquet de Roses,6500,...,10
```
(Réduire la quantité à 10)

Redémarrer SuperShop (Terminal 2).

**Résultat** : Discount plus fort (10% au lieu de 5%)

---

### Scénario 2 : Objectif Marge → Stratégie Premium

Voir `sample_implementation/main_with_multiagent.go.example`

Changer :
```go
Objective: "margin",  // au lieu de "volume"
```

**Résultat** : Si déjà compétitif, garde le prix (pas de discount)

---

## 🔧 Intégration dans le Code

**IMPORTANT** : Le code actuel utilise encore l'ancien agent monolithique.

Pour activer la nouvelle architecture multi-agents :

1. Ouvrez `sample_implementation/main.go`
2. Copiez le code de `main_with_multiagent.go.example`
3. Remplacez la fonction `newMux()` par `newMuxWithMultiAgent()`

Ou suivez le guide complet : `LAUNCH_MULTI_AGENT.md`

---

## ❓ Problèmes Fréquents

### ❌ "Shopping Graph n'est pas accessible"
→ Le Terminal 1 ne tourne pas. Lancez le Shopping Graph.

### ❌ "Merchant sur :8182 non accessible"
→ Les Terminals 2/3/4 ne tournent pas. Lancez les merchants.

### ❌ "AUTO_COMPETE n'a pas été appliqué"
→ Le code n'utilise pas encore la nouvelle architecture.
→ Voir `main_with_multiagent.go.example`

### ⚠️ "Shopping Graph a indexé 0 merchants"
→ Attendez 30 secondes de plus, le polling est toutes les 30s.

---

## 📚 Documentation Complète

- `LAUNCH_MULTI_AGENT.md` - Guide détaillé
- `pkg/merchant/competitive/INTEGRATION.md` - Intégration dans le code
- `docs/adr/003-competitive-pricing-agent.md` - Architecture Decision Record

---

## 🎉 C'est Tout !

En 3 étapes, vous avez testé l'architecture multi-agents qui :

✅ Analyse le marché automatiquement  
✅ Choisit la stratégie selon le contexte  
✅ Bat la concurrence intelligemment  
✅ Explique ses décisions  
✅ Garantit la rentabilité  

**Next step** : Intégrer dans le code pour utiliser en production !
