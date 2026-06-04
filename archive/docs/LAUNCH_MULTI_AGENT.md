# Guide de Lancement - Architecture Multi-Agents

## 🎯 Objectif

Tester l'architecture multi-agents pour le competitive pricing avec le code "AUTO_COMPETE".

---

## ⚡ Lancement Rapide (5 minutes)

### Étape 1 : Lancer le Shopping Graph

Le Shopping Graph est nécessaire pour récupérer les prix concurrents.

```bash
# Terminal 1
cd demo
go run ./cmd/shopping-graph --port 9000
```

Vous devriez voir :
```
Shopping Graph listening on :9000
Polling merchants every 30s...
```

---

### Étape 2 : Lancer 3 Merchants (pour avoir des concurrents)

Chaque merchant doit tourner sur un port différent avec des données différentes.

```bash
# Terminal 2 - Merchant A (SuperShop)
cd sample_implementation
go run . --port 8182 \
  --data-dir ../demo/data/merchant_a \
  --merchant-name "SuperShop"
```

```bash
# Terminal 3 - Merchant B (MegaMart)
cd sample_implementation
go run . --port 8183 \
  --data-dir ../demo/data/merchant_b \
  --merchant-name "MegaMart"
```

```bash
# Terminal 4 - Merchant C (BudgetBuy)
cd sample_implementation
go run . --port 8184 \
  --data-dir ../demo/data/merchant_c \
  --merchant-name "BudgetBuy"
```

Attendez ~30 secondes que le Shopping Graph indexe les 3 merchants.

---

### Étape 3 : Tester sans AUTO_COMPETE (baseline)

Créer un checkout normal chez SuperShop :

```bash
curl -X POST http://localhost:8182/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": "prod_roses_bouquet",
        "quantity": 1
      }
    ]
  }' | jq
```

Notez le prix total (ex: $65.00).

---

### Étape 4 : Tester avec AUTO_COMPETE

Appliquer le code "AUTO_COMPETE" :

```bash
# Remplacez CHECKOUT_ID par l'ID du checkout créé ci-dessus
curl -X PATCH http://localhost:8182/checkout/CHECKOUT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "discount_codes": ["AUTO_COMPETE"]
  }' | jq
```

**Regardez les logs du Terminal 2** pour voir les agents en action :

```
[Orchestrator] Starting competitive pricing analysis for product prod_roses_bouquet
[Orchestrator] Price Intelligence: rank 2/3, lowest: $59.99 (MegaMart)
[Orchestrator] Market Analysis: follower position, stable trend, opportunity: optimize
[Orchestrator] Strategy: balanced, target: $56.99, discount: $8.01, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]
[Orchestrator] ✅ Pricing approved: $56.99 (discount: $8.01, margin: 25%)
```

Le prix devrait maintenant être **inférieur** au concurrent le plus bas !

---

## 🧪 Tests des Différents Scénarios

### Scénario 1 : Stock Normal (Balanced Strategy)

Par défaut, avec un stock normal, la stratégie est "balanced" (beat by 5%).

**Résultat attendu** : Prix concurrent - 5%

---

### Scénario 2 : Stock Bas (Aggressive Strategy)

Modifier le stock dans `demo/data/merchant_a/products.csv` :

```csv
id,name,price,description,image_url,quantity
prod_roses_bouquet,Bouquet de Roses,6500,Magnifique bouquet de roses rouges,https://example.com/roses.jpg,10
```

Redémarrer SuperShop (Terminal 2).

**Résultat attendu** : Prix concurrent - 10% (plus agressif car stock bas)

Logs attendus :
```
[Orchestrator] Strategy: aggressive, target: $53.99
[Orchestrator] Reasoning: ["Low stock (10 units) - clear inventory quickly"]
```

---

### Scénario 3 : Déjà Leader (Premium Strategy)

Si SuperShop a déjà le prix le plus bas, tester avec un produit où MegaMart et BudgetBuy sont plus chers.

**Résultat attendu** : Pas de discount (ou discount minimal 2%)

Logs attendus :
```
[Orchestrator] Strategy: premium
[Orchestrator] Reasoning: ["Already market leader - maximize margin"]
```

---

### Scénario 4 : Guerre des Prix (Match Strategy)

Pour simuler une guerre des prix, il faudrait que les prix baissent rapidement dans l'historique.

(Plus complexe, nécessite modification du code pour injecter un historique fictif)

---

## 🎛️ Configurer le Comportement

### Changer l'Objectif Business

Dans `main_with_multiagent.go.example`, modifier :

```go
businessConfig := models.BusinessConfig{
    Objective:      "margin",  // ← Changer de "volume" à "margin"
    StockThreshold: 20,
    BrandPosition:  "premium", // ← Changer à "premium"
    MinMargin:      15,        // ← Augmenter la marge minimum
    CostPercent:    60,
}
```

**Avec `Objective: "margin"`** :
- Si déjà compétitif → garde le prix (premium strategy)
- Moins agressif en général

**Avec `BrandPosition: "premium"`** :
- Accepte d'être un peu plus cher
- Focus sur la marge

---

## 📊 Vérifier les Résultats

### 1. Comparer les Prix

```bash
# Prix chez les 3 merchants
curl http://localhost:8182/api/products | jq '.[] | {name, price}'
curl http://localhost:8183/api/products | jq '.[] | {name, price}'
curl http://localhost:8184/api/products | jq '.[] | {name, price}'
```

### 2. Vérifier le Shopping Graph

```bash
curl -X POST http://localhost:9000/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "roses",
    "limit": 10
  }' | jq
```

Vous devriez voir les 3 merchants avec leurs prix.

---

## 🐛 Dépannage

### Le Shopping Graph ne trouve pas les produits

**Problème** : Les merchants ne sont pas encore indexés.

**Solution** : Attendez 30 secondes après avoir lancé les merchants.

Vérifiez :
```bash
curl http://localhost:9000/search -d '{"query":"roses"}' | jq '.total'
```

Si `total: 0`, les merchants ne sont pas indexés.

---

### Erreur "shopping graph search failed"

**Problème** : Le Shopping Graph n'est pas accessible.

**Solution** : Vérifiez que le Shopping Graph tourne sur le port 9000 :
```bash
curl http://localhost:9000/health
```

---

### Le discount AUTO_COMPETE n'est pas appliqué

**Problème** : Le code actuel utilise l'ancien agent monolithique.

**Solution** : Vous devez intégrer la nouvelle architecture en modifiant `main.go`.

Voir la section suivante.

---

## 🔧 Intégration dans le Code (Avancé)

### Option A : Utiliser l'Exemple

1. Copiez le contenu de `sample_implementation/main_with_multiagent.go.example`
2. Modifiez `sample_implementation/main.go` :
   - Ajoutez les imports
   - Remplacez `newMux()` par `newMuxWithMultiAgent()`

### Option B : Intégration Manuelle

Voir le guide complet dans `pkg/merchant/competitive/INTEGRATION.md`

---

## 📈 Scénario Complet de Démo

### Préparation

1. Lancez Shopping Graph (port 9000)
2. Lancez 3 merchants (ports 8182, 8183, 8184)
3. Attendez 30s pour indexation

### Étape 1 : Vérifier les prix de base

```bash
# SuperShop : Roses $65.00
curl http://localhost:8182/api/products | jq '.[] | select(.id=="prod_roses_bouquet")'

# MegaMart : Roses $59.99
curl http://localhost:8183/api/products | jq '.[] | select(.id=="prod_roses_bouquet")'

# BudgetBuy : Roses $70.00
curl http://localhost:8184/api/products | jq '.[] | select(.id=="prod_roses_bouquet")'
```

**Résultat** : MegaMart est le moins cher ($59.99)

---

### Étape 2 : Créer checkout chez SuperShop SANS AUTO_COMPETE

```bash
CHECKOUT=$(curl -s -X POST http://localhost:8182/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "prod_roses_bouquet", "quantity": 1}]
  }' | jq -r '.id')

echo "Checkout ID: $CHECKOUT"

# Voir le prix
curl http://localhost:8182/checkout/$CHECKOUT | jq '.totals'
```

**Résultat** : Total = $65.00 (prix de base SuperShop)

---

### Étape 3 : Appliquer AUTO_COMPETE

```bash
curl -X PATCH http://localhost:8182/checkout/$CHECKOUT \
  -H "Content-Type: application/json" \
  -d '{
    "discount_codes": ["AUTO_COMPETE"]
  }' | jq '.totals'
```

**Regardez les logs du merchant SuperShop !**

**Résultat attendu** :
- Discount appliqué : ~$8.01
- Nouveau total : ~$56.99
- C'est **moins cher** que MegaMart ($59.99) !

---

### Étape 4 : Vérifier le raisonnement

Dans les logs, vous verrez :

```
[Orchestrator] Price Intelligence: rank 2/3, lowest: $59.99 (MegaMart)
[Orchestrator] Strategy: balanced, target: $56.99, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]
[Orchestrator] ✅ Pricing approved: $56.99 (discount: $8.01, margin: 25%)
```

**Explication** :
- Agent 1 : Trouvé MegaMart à $59.99
- Agent 2 : Position "follower", marché stable
- Agent 3 : Recommande "balanced" (beat by 5%)
- Agent 4 : Marge 25% ✓, approuvé

---

## 🎉 Résultat Final

Vous avez maintenant un système de pricing intelligent qui :

1. **S'adapte au contexte** (stock, objectif, position marché)
2. **Bat la concurrence** automatiquement
3. **Explique ses décisions** dans les logs
4. **Garantit la rentabilité** (marge minimum)

---

## 📝 Prochaines Étapes

1. **Modifier le stock** dans merchant_a pour voir la stratégie aggressive
2. **Changer l'objectif** de "volume" à "margin" pour voir le comportement premium
3. **Créer ADR 004** documentant les évolutions de l'architecture multi-agents
4. **Ajouter tests** pour chaque agent

---

## 📚 Ressources

- `pkg/merchant/competitive/INTEGRATION.md` - Guide d'intégration détaillé
- `docs/adr/003-competitive-pricing-agent.md` - ADR du pricing agent
- `sample_implementation/main_with_multiagent.go.example` - Code d'exemple
