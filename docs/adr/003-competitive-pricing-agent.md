# ADR 003 : Agent de Pricing Dynamique Compétitif

- **Date** : 2026-03-30
- **Statut** : Accepté
- **Décideurs** : Olivier Wulveryck
- **Lié à** : ADR 001 (Architecture Multi-Agent), ADR 002 (Multi-Transport)

## Contexte

L'architecture multi-agent (ADR 001) et l'Arena mode créent un environnement de **compétition entre merchants**. Dans ce contexte, les merchants doivent pouvoir ajuster leurs prix dynamiquement pour rester compétitifs et gagner des ventes.

### Problème

**Comment un merchant peut-il optimiser son pricing en temps réel pour battre la concurrence tout en maintenant sa rentabilité ?**

### Exigences fonctionnelles

- **Pricing compétitif** : Détecter les prix concurrents et ajuster automatiquement
- **Maintien de la marge** : Ne jamais descendre sous un seuil de rentabilité minimum
- **Stratégies configurables** : Différentes approches selon le contexte business
- **Intégration transparente** : S'intègre au système de discount existant sans refonte

### Exigences non-fonctionnelles

- ✅ **Performance** : Latence < 1s pour calcul de discount compétitif
- ✅ **Résilience** : Dégradation gracieuse si Shopping Graph indisponible
- ✅ **Auditabilité** : Logs de toutes les décisions de pricing
- ✅ **Simplicité** : Injection via interface existante (`discount.DiscountLookup`)

### Scénario Utilisateur

```
1. Client Agent cherche "wireless headphones"
2. Shopping Graph retourne :
   - SuperShop : $89.99
   - MegaMart : $84.99 ← Lowest
   - BudgetBuy : $95.00

3. Client crée checkout chez SuperShop avec code "AUTO_COMPETE"
4. Competitive Pricing Agent :
   - Détecte code spécial "AUTO_COMPETE"
   - Query Shopping Graph → lowest = $84.99 (MegaMart)
   - Calcule discount pour battre : $89.99 → $80.75 (5% sous MegaMart)
   - Valide marge minimale : (80.75 - 60) / 80.75 = 25% ✓ (min 10%)
   - Applique discount de $9.24

5. Client compare totaux :
   - SuperShop : $80.75 (avec AUTO_COMPETE)
   - MegaMart : $84.99
   → Achète chez SuperShop !
```

---

## Décision

Implémenter un **Competitive Pricing Agent** qui calcule des discounts dynamiques basés sur les prix concurrents récupérés via le Shopping Graph.

### Architecture

```
┌────────────────────────────────────────────────────┐
│            Checkout Update Request                 │
│      (discount_code: "AUTO_COMPETE")               │
└───────────────────┬────────────────────────────────┘
                    │
     ┌──────────────▼──────────────────┐
     │  Merchant.UpdateCheckout()      │
     └──────────────┬──────────────────┘
                    │
     ┌──────────────▼──────────────────────────────┐
     │  CompetitivePricingAgent                    │
     │  (implements discount.DiscountLookup)       │
     │                                              │
     │  1. Detect "AUTO_COMPETE" code              │
     │  2. For each line item:                     │
     │     ├─ Query ShoppingGraphClient            │
     │     ├─ Calculate discount (strategy)        │
     │     ├─ Validate margin constraint           │
     │     └─ Sum total discount                   │
     │  3. Return dynamic discount                 │
     └──────────────┬──────────────────────────────┘
                    │
     ┌──────────────▼──────────────────┐
     │  ShoppingGraphClient             │
     │  • HTTP client (1s timeout)      │
     │  • Price cache (10s TTL)         │
     │  • GET /search                   │
     └──────────────┬──────────────────┘
                    │
     ┌──────────────▼──────────────────┐
     │  Shopping Graph (:9000)          │
     │  • Indexes all merchant prices   │
     │  • Returns sorted results        │
     └──────────────────────────────────┘
```

### Composants

## 1. CompetitivePricingAgent

**Responsabilité** : Calculer des discounts dynamiques basés sur les prix concurrents

**Implémentation** : `pkg/merchant/competitive/agent.go` (~190 LOC)

**Interface** :
```go
type CompetitivePricingAgent struct {
    baseData      discount.DiscountLookup  // Static discounts fallback
    competitorAPI CompetitorPriceSource    // Shopping Graph client
    config        Config                    // Strategy, margins, etc.
    merchantID    string
}

// Implements discount.DiscountLookup
func (a *CompetitivePricingAgent) FindDiscountByCode(code string) *discount.Discount
func (a *CompetitivePricingAgent) ApplyDiscountsWithContext(
    codes []string,
    lineItems []model.LineItem,
) []discount.AppliedDiscount
```

**Stratégies de Pricing** :

| Stratégie | Comportement | Use Case |
|-----------|--------------|----------|
| `StrategyMatchPrice` | Match le prix concurrent exactement | Guerre des prix pure |
| `StrategyBeatPrice` | Bat de X% ou $Y minimum | Positionnement agressif |
| `StrategyAutoDiscount` | Bat de 5% ou $0.25 (auto) | Mode démo/default |

**Configuration** :
```go
type Config struct {
    Strategy          PricingStrategy  // Match, Beat, Auto
    BeatByPercent     int             // 5% par défaut
    BeatByMinAmount   int             // $0.25 minimum
    MinMarginPercent  int             // 10% minimum
    CostPricePercent  int             // 60% du prix = coût
}
```

**Algorithme** :

```go
// Pour chaque ligne item
func (a *CompetitivePricingAgent) calculateItemDiscount(item model.LineItem) DiscountCalculation {
    // 1. Déterminer notre prix unitaire
    ourUnitPrice := item.Totals["total"] / item.Quantity
    
    // 2. Query Shopping Graph
    competitorPrice, competitorID, err := a.competitorAPI.GetLowestPrice(productID)
    if err != nil {
        return noDiscount("no competitor data")
    }
    
    // 3. Early exit si déjà compétitif
    if ourUnitPrice <= competitorPrice {
        return noDiscount("already cheaper")
    }
    
    // 4. Calculer discount selon stratégie
    discount := calculateStrategyDiscount(ourUnitPrice, competitorPrice)
    
    // 5. Valider marge minimale
    finalPrice := ourUnitPrice - discount
    margin := (finalPrice - costPrice) / finalPrice
    if margin < a.config.MinMarginPercent {
        return rejected("margin too low")
    }
    
    // 6. Appliquer
    return applied(discount)
}
```

---

## 2. ShoppingGraphClient

**Responsabilité** : Récupérer les prix concurrents depuis le Shopping Graph

**Implémentation** : `pkg/merchant/competitive/shoppinggraph.go` (~192 LOC)

**Interface** :
```go
type CompetitorPriceSource interface {
    GetLowestPrice(productID string) (price int, merchantID string, error)
    GetCompetitorPrices(productID string) ([]CompetitorPrice, error)
}
```

**Features** :
- **HTTP client** : Timeout 1s (fail-fast)
- **Cache in-memory** : TTL 10s (réduit les appels Shopping Graph)
- **Search POST /search** : Query par product ID, limite 10 résultats
- **Filter in-stock** : Ignore les produits en rupture
- **Thread-safe** : sync.RWMutex sur le cache

**Exemple requête** :
```json
POST http://localhost:9000/search
{
  "query": "prod_headphones_wireless",
  "limit": 10
}

Response:
{
  "results": [
    {
      "merchant_id": "merchant_b",
      "merchant_name": "MegaMart",
      "product_id": "prod_headphones_wireless",
      "price": 8499,
      "in_stock": true
    },
    {
      "merchant_id": "merchant_a",
      "merchant_name": "SuperShop",
      "price": 8999,
      "in_stock": true
    }
  ],
  "total": 2
}
```

**Cache Implementation** :
```go
type priceCache struct {
    mu      sync.RWMutex
    entries map[string]cachedPrice
    ttl     time.Duration  // 10s
}

func (pc *priceCache) get(productID string) (cachedPrice, bool) {
    if time.Since(entry.Timestamp) > pc.ttl {
        return cachedPrice{}, false  // Expired
    }
    return entry, true
}
```

---

## 3. Competitive Intelligence UI

**Responsabilité** : Dashboard pour visualiser et appliquer les recommandations de prix

**Implémentation** : `demo/cmd/arena/competitive_intel.go` (~180 LOC)

**Endpoint** : `GET /api/competitive-intel`

**Response** :
```json
{
  "our_price": 8999,
  "our_price_display": "$89.99",
  "lowest_price": 8499,
  "lowest_price_by": "MegaMart",
  "competitors": [
    {
      "merchant_name": "MegaMart",
      "price": 8499,
      "price_display": "$84.99",
      "is_us": false
    },
    {
      "merchant_name": "SuperShop",
      "price": 8999,
      "price_display": "$89.99",
      "is_us": true
    }
  ],
  "recommended_price": 8075,
  "recommended_price_display": "$80.75",
  "price_difference": -924,
  "margin_percent": 25,
  "would_win": true,
  "message": "💡 Lower to $80.75 to beat MegaMart and win sales!"
}
```

**UI Features** :
- **Auto-refresh** : Toutes les 10 secondes
- **Visual feedback** : Prix concurrent le plus bas en surbrillance
- **One-click apply** : Bouton "Appliquer ce prix" ajuste le slider
- **Margin validation** : Affiche la marge après ajustement

---

## Alternatives Considérées

### Alternative 1 : Hardcoded Discount Codes

**Description** : Codes promo statiques définis manuellement

**Pour** :
- ✅ Simple à implémenter
- ✅ Pas de dépendances externes
- ✅ Prévisible

**Contre** :
- ❌ Pas d'adaptation aux concurrents
- ❌ Nécessite updates manuelles
- ❌ Pas de garantie de compétitivité

**Verdict** : ❌ Rejeté. Ne résout pas le problème de pricing dynamique.

---

### Alternative 2 : API Externe de Pricing (ex: Price2Spy)

**Description** : Service tiers spécialisé dans le price tracking

**Pour** :
- ✅ Données riches (historique, tendances)
- ✅ Monitoring actif 24/7
- ✅ Pas besoin de maintenir l'infrastructure

**Contre** :
- ❌ Coût ($$ par requête)
- ❌ Latence réseau externe
- ❌ Dépendance tiers (SLA, uptime)
- ❌ Over-engineering pour une démo

**Verdict** : ❌ Rejeté. Trop complexe et coûteux pour notre use case.

---

### Alternative 3 : Machine Learning Model

**Description** : Modèle ML prédisant le prix optimal basé sur historique

**Pour** :
- ✅ Prédictions sophistiquées
- ✅ Apprend des patterns
- ✅ Peut anticiper les tendances

**Contre** :
- ❌ Nécessite dataset d'entraînement
- ❌ Complexité infrastructure (serving model)
- ❌ Latence inference (~100-500ms)
- ❌ Difficult à debugger/expliquer

**Verdict** : ❌ Rejeté. Over-engineering. Règles simples suffisent.

---

### Alternative 4 : Competitive Pricing Agent ✅

**Description** : Agent intégré qui query le Shopping Graph et applique des stratégies configurables

**Pour** :
- ✅ Temps réel (< 1s latence totale)
- ✅ Utilise l'infrastructure existante (Shopping Graph)
- ✅ Stratégies explicables (Match, Beat, Auto)
- ✅ Injection via interface existante (discount.DiscountLookup)
- ✅ Validations business (marge minimale)

**Contre** :
- ❌ Dépendance au Shopping Graph
- ❌ Cache nécessaire pour performance
- ❌ Pas de sophistication ML

**Verdict** : ✅ **Choisi**. Balance simplicité/efficacité/performance.

---

## Trade-offs

### Positifs

✅ **Intégration transparente** : Implémente `discount.DiscountLookup` → zéro refonte  
✅ **Performance** : Cache 10s + timeout 1s → latence < 1s total  
✅ **Sécurité business** : Validation marge minimale → pas de ventes à perte  
✅ **Flexibilité** : 3 stratégies configurables selon contexte  
✅ **Auditabilité** : Logs détaillés de chaque décision  
✅ **Resilience** : Fallback sur static discounts si Shopping Graph down  

### Négatifs

❌ **Dépendance Shopping Graph** : Si down → pas de pricing compétitif  
❌ **Cache staleness** : Prix concurrents peuvent changer pendant 10s TTL  
❌ **Simplicité algorithme** : Pas de ML/prédiction sophistiquée  
❌ **Complexité** : ~700 LOC vs discount statique (~50 LOC)  

### Risques et Mitigations

**Risque 1 : Shopping Graph indisponible**
- Mitigation : Fallback sur `baseData` (static discounts)
- Mitigation : Timeout 1s (fail-fast)
- Impact : Perte de compétitivité temporaire, pas de crash

**Risque 2 : Race condition (2 merchants baissent simultanément)**
- Mitigation : Cache 10s réduit la fréquence
- Mitigation : Validation marge minimale empêche spirale
- Impact : Possible "price war" temporaire

**Risque 3 : Exploitation du code "AUTO_COMPETE"**
- Mitigation : Validation marge minimale (hard floor)
- Mitigation : Logs auditable de chaque discount appliqué
- Impact : Limité par la marge minimale configurée

---

## Conséquences

### Impact Architecture

**Commit `700eeef` (30 mars 2026)** : Ajout Competitive Pricing
- +706 lignes (`pkg/merchant/competitive/`)
- +180 lignes (UI `demo/cmd/arena/competitive_intel.go`)
- Aucune modification du core merchant (ADR 002 validé)

**Pattern validé** : Injection via interface
```go
// Before
merchant := newSimpleMerchant(catalog, shopData, staticDiscountLookup, ...)

// After
competitiveAgent := competitive.NewCompetitivePricingAgent(
    staticDiscountLookup,      // Fallback
    shoppingGraphClient,        // Prix concurrents
    competitive.StrategyBeatPrice,
    10,  // 10% marge minimum
)
merchant := newSimpleMerchant(catalog, shopData, competitiveAgent, ...)
```

### Code Example : Intégration

```go
// 1. Créer le Shopping Graph client
sgClient := competitive.NewShoppingGraphClient("http://localhost:9000")

// 2. Créer l'agent avec stratégie
agent := competitive.NewCompetitivePricingAgent(
    shopData,                          // Base discount data
    sgClient,                           // Competitor price source
    competitive.StrategyBeatPrice,      // Beat by %
    10,                                 // 10% min margin
)
agent.SetMerchantID(merchantID)
agent.SetBeatByPercent(5)              // Beat by 5%
agent.SetCostPricePercent(60)          // Cost = 60% of price

// 3. Injecter dans le merchant
merchant := newSimpleMerchant(
    catalog,
    shopData,
    agent,  // ← Competitive pricing enabled
    fulfillmentData,
    paymentProcessor,
)

// 4. Client utilise "AUTO_COMPETE"
checkout := merchant.UpdateCheckout(ctx, model.CheckoutRequest{
    DiscountCodes: []string{"AUTO_COMPETE"},
    Items: [...],
})
// → Discount dynamiquement calculé !
```

### Métriques

| Métrique | Valeur |
|----------|--------|
| **LOC Agent** | ~706 (agent 190 + client 192 + types 263 + doc 61) |
| **LOC UI** | ~180 (competitive_intel.go) |
| **Latence p50** | ~300ms (cache hit : ~1ms, cache miss : ~300ms) |
| **Latence p99** | ~1000ms (timeout Shopping Graph) |
| **Cache hit rate** | ~85% (10s TTL, refresh UI 10s) |
| **Stratégies** | 3 (Match, Beat, Auto) |

---

## Évolutions Post-Implémentation

### Retours Terrain (Mai 2026)

**Ce qui fonctionne bien** :
- Intégration via `discount.DiscountLookup` est seamless
- Cache 10s équilibre fraîcheur/performance
- UI Intelligence Compétitive très appréciée en démo
- Validation marge empêche les erreurs business

**Axes d'amélioration** :
- **Cache partagé** : Actuellement par merchant. Partager entre merchants ?
- **Historique prix** : Pas de tracking des changements concurrents
- **Alertes** : Pas de notification si concurrent baisse drastiquement
- **A/B Testing** : Pas de framework pour tester stratégies

### Évolutions Possibles

**1. Prix Différenciés par Région**
```go
type RegionalPricingAgent struct {
    agents map[string]*CompetitivePricingAgent  // region → agent
}
```

**2. Pricing Basé sur le Stock**
```go
// Si stock bas → prix plus agressif pour vendre rapidement
if stockLevel < threshold {
    agent.SetBeatByPercent(10)  // Beat by 10% instead of 5%
}
```

**3. Time-Based Pricing**
```go
// Happy hour : prix plus bas 18h-20h
if isHappyHour() {
    agent.SetBeatByPercent(15)
}
```

---

## Tableau Comparatif Stratégies

| Stratégie | Notre Prix | Concurrent | Résultat | Marge | Use Case |
|-----------|-----------|-----------|----------|-------|----------|
| **Match** | $89.99 | $84.99 | **$84.99** | 20% | Guerre prix |
| **Beat 5%** | $89.99 | $84.99 | **$80.75** | 15% | Agressif |
| **Beat 10%** | $89.99 | $84.99 | **$76.50** | 10% | Très agressif |
| **Auto** | $89.99 | $84.99 | **$84.14** | 17% | Défaut démo |

**Coût** : $60 (60% du prix base)  
**Marge min** : 10% configurée

---

## Références

### Architecture
- Shopping Graph : Service d'agrégation cross-merchant (ADR 001)
- Discount System : `pkg/merchant/discount/` (interface DiscountLookup)

### ADR Liés
- [ADR 001](001-multi-agent-shopping-architecture.md) : Architecture Multi-Agent (Shopping Graph)
- [ADR 002](002-multi-transport-architecture.md) : Multi-Transport (injection pattern)

### Commits Clés
- `700eeef` (2026-03-30) : Add dynamic pricing algorithms with auto bid and discount management
- `4229392` : Add merchant dashboard UX improvements

### Documentation
- `COMPETITIVE_PRICING.md` : Guide utilisateur de l'intelligence compétitive
- `pkg/merchant/competitive/doc.go` : Documentation package

### Interfaces Utilisées
```go
// pkg/merchant/discount/discount.go
type DiscountLookup interface {
    FindDiscountByCode(code string) *Discount
    ApplyDiscountsWithContext(codes []string, items []model.LineItem) []AppliedDiscount
}
```
