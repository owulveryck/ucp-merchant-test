# Référence : Les 4 Agents Arena

## Vue d'ensemble

| Agent | Nom complet | Rôle | Département |
|-------|-------------|------|-------------|
| 🕵️ | L'Espion | Price Intelligence | Intelligence Compétitive |
| 📊 | L'Analyste | Market Analysis | Analyse Stratégique |
| 🎯 | Le Stratège | Strategy Recommender | Stratégie Commerciale |
| ✅ | Le Contrôleur | Margin Validator | Contrôle Financier |

---

## 🕵️ L'Espion (Price Intelligence)

**Fonction** : `GetCompetitorPrices(product_id)`

**Entrée** :
- `product_id` : ID du produit (ex: "laptop")

**Sortie** :
```json
{
  "product_id": "laptop",
  "competitors": [
    {"name": "Concurrent A", "price": 100000},
    {"name": "Concurrent B", "price": 105000},
    {"name": "Concurrent C", "price": 95000}
  ],
  "min_price": 95000,
  "max_price": 105000,
  "avg_price": 100000
}
```

**Temps d'exécution** : ~50ms

---

## 📊 L'Analyste (Market Analysis)

**Fonction** : `AnalyzeMarket(product_id, proposed_price, competitor_prices)`

**Entrée** :
- `product_id` : ID du produit
- `proposed_price` : Prix que vous envisagez (en centimes)
- `competitor_prices` : Liste des prix concurrents

**Sortie** :
```json
{
  "market_position": 2,
  "total_competitors": 4,
  "is_competitive": true,
  "price_range": {
    "min": 95000,
    "max": 105000
  },
  "recommendation": "Good positioning"
}
```

**Temps d'exécution** : ~30ms

---

## 🎯 Le Stratège (Strategy Recommender)

**Fonction** : `RecommendStrategy(market_analysis, customer_tier)`

**Entrée** :
- `market_analysis` : Résultat de l'Analyste
- `customer_tier` : Tier du client (standard/silver/gold/premium)

**Sortie** :
```json
{
  "strategy": "match_lowest",
  "recommended_price": 95000,
  "rationale": "Align with lowest competitor to maximize win chance",
  "win_probability": 0.85,
  "alternative_strategies": [
    {"name": "premium", "price": 105000, "win_probability": 0.30},
    {"name": "undercut", "price": 94000, "win_probability": 0.95}
  ]
}
```

**Stratégies disponibles** :
- `match_lowest` : S'aligner sur le concurrent le moins cher
- `undercut` : Battre le concurrent le moins cher de 1-2%
- `premium` : Prix haut de gamme (pour clients Premium)
- `average` : Prix moyen du marché

**Temps d'exécution** : ~40ms

---

## ✅ Le Contrôleur (Margin Validator)

**Fonction** : `ValidateMargin(product_id, selling_price, cost_price)`

**Entrée** :
- `product_id` : ID du produit
- `selling_price` : Prix de vente proposé (en centimes)
- `cost_price` : Coût d'achat du produit (en centimes)

**Sortie** :
```json
{
  "is_valid": true,
  "margin_amount": 15000,
  "margin_percent": 15.8,
  "min_required_margin": 10.0,
  "decision": "APPROVED",
  "reason": "Margin 15.8% exceeds minimum 10%"
}
```

**Règles de validation** :
- ✅ Marge ≥ 10% : APPROUVÉ
- ❌ Marge < 10% : REJETÉ
- ⚠️ Marge < 5% : REJETÉ avec alerte

**Temps d'exécution** : ~20ms

---

## Flux complet

```
Produit: laptop, Prix initial: $1000

1. 🕵️ L'Espion
   Input:  product_id = "laptop"
   Output: min=$950, max=$1050, avg=$1000
   Temps:  50ms

2. 📊 L'Analyste
   Input:  proposed=$1000, competitors=[$950,$1000,$1050]
   Output: position=2/4, competitive=true
   Temps:  30ms

3. 🎯 Le Stratège
   Input:  market_analysis + customer_tier="gold"
   Output: strategy="match_lowest", price=$950
   Temps:  40ms

4. ✅ Le Contrôleur
   Input:  selling=$950, cost=$800
   Output: margin=15.8%, APPROVED
   Temps:  20ms

TOTAL: 140ms pour une décision complète
```

---

## Coûts par produit (référence)

| Produit | Coût d'achat | Marge minimum | Prix minimum |
|---------|--------------|---------------|--------------|
| laptop | $800 | 10% = $80 | $880 |
| mouse | $20 | 10% = $2 | $22 |
| keyboard | $55 | 10% = $5.50 | $60.50 |
| monitor | $280 | 10% = $28 | $308 |

---

## Retour

[Comprendre les 4 agents](../explanation/les-4-agents.md)
