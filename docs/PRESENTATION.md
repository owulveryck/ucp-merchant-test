---
marp: true
theme: default
paginate: true
---

# UCP Merchant Test
## Système de Pricing Intelligent Multi-Agents

**Stage OCTO - Juin 2026**

---

## Vue d'ensemble

**Serveur marchand Go** implémentant :
- Universal Commerce Protocol (UCP)
- Model Context Protocol (MCP)
- Agent-to-Agent (A2A)

**3 niveaux d'architecture** :
1. Base UCP conforme (60 tests ✅)
2. Shopping multi-agents (démo Gemini)
3. Arena compétitive (5 marchands IA)

---

## Objectif Principal

Démontrer l'utilisation de **systèmes multi-agents** pour :

- **Pricing compétitif** en temps réel
- **Shopping intelligent** cross-merchant
- **Décisions autonomes** avec contraintes business
- **Interopérabilité** protocoles standards (UCP/MCP/A2A)

---

## Couche 1 : UCP Merchant Base

### Conformance Complète
- ✅ **60 tests UCP** passés (13 fichiers)
- Transports : **REST** + **MCP** + **A2A**
- Checkout lifecycle complet
- Gestion : commandes, paiements, shipping, discounts

### Dataset Flower Shop
- 6 produits (1 out-of-stock)
- 3 clients, codes promo (10OFF, WELCOME20, FIXED500)
- Free shipping conditionnel

---

## Couche 2 : Multi-Agent Shopping Demo

```
Client Agent (Gemini)
    ↓
Shopping Graph (:9000)
    ↓
    ├─→ SuperShop  (:8182) - 6 produits, SAVE10/WELCOME15
    ├─→ MegaMart   (:8183) - 5 produits, MEGA10
    └─→ BudgetBuy  (:8184) - 5 produits, BUDGET20/SAVE5
         ↓
    Obs Hub (:9002) - Monitoring SSE
```

**Agent Gemini** : recherche, compare, applique promos, choisit le moins cher

---

## Couche 3 : Competitive Pricing Arena

### Architecture 3-Agents Orchestrée

```
Agent 1 (Vendor Orchestrator)
    ↓
    ├─→ Agent 2 (Customer Growth)
    │   └─→ Fidélisation, discount VIP
    │
    └─→ Agent 3 (Competitiveness)
        └─→ Price Intelligence + Market Analysis
            └─→ Wraps 4-agent system
```

---

## Arena - 5 Marchands en Compétition

### Participants
- MegaStore
- PrixCassés
- SuperDeals
- TopPrix
- MonMarchand

### Stratégies
- `vip_retention` - Fidéliser clients premium
- `balanced` - Équilibre marge/compétitivité
- `premium` - Position leader
- `match_market` - Suivre le marché

---

## Résultats Démo Arena (5 juin 16h)

### Produit : Casque Audio

**Premier pricing** (16h03)
- Prix initial : $62.15 → **$52.30** (marge 4%)
- Position : 2/5 → **1/5** (leader)
- Stratégie : VIP retention (-10% fidélité)

**Second pricing** (16h05) - Guerre des prix !
- Prix : $52.30 → **$46.13** (marge **-8%** !)
- ⚠️ **Marge sacrifiée** pour garantir victoire
- Position maintenue : **1/5**

---

## Architecture Technique

```
pkg/
├── merchant/
│   ├── transport/        # REST/MCP/A2A
│   ├── discount/         # Codes promo
│   ├── fulfillment/      # Shipping
│   └── pricing/          # Calcul totaux
├── pricing-simple/       # Agents basiques
├── pricing-intelligent/  # 4-agent system
├── pricing-unified/      # 3-agent orchestré ✨
├── model/                # Types UCP
└── auth/                 # OAuth2
```

---

## Architecture Technique (suite)

```
demo/
├── cmd/
│   ├── shopping-graph/   # Index cross-merchant
│   ├── client/           # Agent Gemini
│   ├── obs-hub/          # Dashboard monitoring
│   └── arena/            # Compétition 5 marchands
├── data/                 # Catalogs merchants
└── scripts/              # Automation
```

---

## 10 ADRs Documentés

### Pricing Intelligence (Mai 2026)
- **0001** - Architecture multi-agents
- **0002** - Victoire avant marge parfaite
- **0003** - Détection codes promo

### Interface & UX (Juin 2026)
- **0004** - Architecture 3-agents orchestrée
- **0005** - Agent acheteur intégré
- **0006** - Messages détaillés décisions
- **0007** - Scénario challenge concurrents

---

## ADRs Infrastructure

- **0008** - Multi-agent shopping architecture
- **0009** - Multi-transport (REST/MCP/A2A)
- **0010** - Competitive pricing agent

**Format** : Template MADR
- Contexte et Problème
- Facteurs de Décision
- Options Considérées
- Décision avec Conséquences

---

## Points Clés Arena

### Comportement Observé
- **Ultra-agressif** : accepte marges négatives pour gagner
- **Temps réel** : ajustements instantanés via SSE
- **Transparent** : chaque agent explique ses décisions

### Problématique Identifiée
- Stratégie "victoire à tout prix" peut sacrifier rentabilité
- Nécessité de gardes-fous (marge minimale absolue)
- Trade-off acquisition vs. profitabilité

---

## Technologies Utilisées

### Backend
- **Go 1.24** - Serveur UCP
- **Gemini (Vertex AI)** - Client agent
- **SSE** - Real-time events

### Protocoles
- **UCP 2026-01-11** - Universal Commerce
- **MCP** - Model Context Protocol (JSON-RPC)
- **A2A** - Agent-to-Agent

### Patterns
- Multi-agent orchestration
- Event-driven architecture
- Strategy pattern (pricing)

---

## Démos Disponibles

### 1. Conformance UCP
```bash
go run ./sample_implementation --port 8182 \
  --data-dir test_data/flower_shop
```

### 2. Shopping Multi-Agents
```bash
demo/scripts/run_demo.sh
# Puis : demo/bin/client --graph-url http://localhost:9000
```

### 3. Arena Compétitive
```bash
demo/bin/arena
# Interface : http://localhost:8888
```

---

## Use Cases Démontrés

1. **E-commerce Standard**
   - Serveur UCP conforme, prêt production

2. **Shopping Intelligence**
   - Agent autonome qui optimise achats cross-merchant

3. **Dynamic Pricing**
   - 5 marchands IA s'affrontent en temps réel
   - Décisions transparentes et observables

---

## Résultats & Apprentissages

### Réussites
- ✅ 60 tests UCP conformance
- ✅ Multi-protocoles (REST/MCP/A2A)
- ✅ Agents autonomes fonctionnels
- ✅ Architecture extensible

### Défis
- ⚠️ Balance compétitivité/rentabilité
- ⚠️ Comportements émergents imprévisibles
- ⚠️ Besoin de règles business strictes

---

## Prochaines Étapes

### Court Terme
- Ajouter seuils de marge absolus
- Implémenter mode "defensive" (protéger marge)
- Tests A/B stratégies pricing

### Moyen Terme
- ML pour prédiction comportements concurrents
- Support multi-devises
- Optimisation latence décisions agents

### Long Terme
- Multi-marketplace (Amazon, eBay, etc.)
- Agent négociateur B2B

---

## Architecture Decision - Highlight

### ADR-0002 : Victoire Avant Marge

**Contexte** : Deux objectifs conflictuels
- Maximiser marge
- Gagner la vente

**Décision** : Priorité à la victoire
- Agent accepte marges réduites
- Marge minimale = 10% (mais contournable si nécessaire)

**Conséquence** : Arena démo → marge -8% !

---

## Démo Live - Arena Timeline

```
16h02:22 - Arena démarrage
16h02:25 - 5 marchands initialisés (système 3-agents)
16h03:40 - Premier pricing casque_audio
           $62.15 → $52.30 (marge 4%)
16h05:35 - Ajustement compétitif
           $52.30 → $46.13 (marge -8%)
           Position: 1/5 (leader)
16h28:29 - MonMarchand quitte l'arène
```

**Observation** : Course vers le bas en 2 minutes

---

## Monitoring & Observabilité

### Events SSE
- `vendor_decision` - Décisions Agent 1
- `customer_analysis` - Analyse Agent 2
- `competitiveness_check` - Analyse Agent 3
- `price_update` - Nouveau prix publié

### Logs Détaillés
```
[Agent Vendeur] → Consultation Agent 2
[Agent Customer Growth] Decision: Tier=gold, Discount=10%
[Agent Compétitivité] Position 1/5 - Prix recommandé: $51.25
[Orchestrator] ⚠️ Marge réduite: 2% pour GAGNER
```

---

## Code Highlight - Agent Orchestration

```go
// Agent 1 orchestrates agents 2 & 3
func (o *Orchestrator) ProposePrice(ctx context.Context, 
    req PriceRequest) (PriceDecision, error) {
    
    // Consult Agent 2 (Customer Growth)
    customerDecision := o.customerGrowth.Analyze(req)
    
    // Consult Agent 3 (Competitiveness)
    competitiveDecision := o.competitiveness.Analyze(req)
    
    // Vendor makes final decision
    return o.synthesize(customerDecision, competitiveDecision)
}
```

---

## Shopping Graph - Cross-Merchant Search

```go
// Jaccard similarity matching
type SearchResult struct {
    MerchantID    string
    MerchantName  string
    ProductID     string
    Title         string
    Price         int64
    DiscountHints []string
    InStock       bool
}

// Returns ranked results from all merchants
func (sg *ShoppingGraph) Search(query string) []SearchResult
```

---

## Stats Projet

### Code
- **~15,000 lignes Go**
- 17 packages principaux
- 10 ADRs documentés

### Tests
- 60 tests conformance UCP
- Tests unitaires Go
- Scénarios démo automatisés

### Démos
- 3 modes complets
- 5 marchands configurables
- Dashboard temps réel

---

## Points de Différenciation

### vs. Pricing Traditionnel
- ✅ Décisions autonomes temps réel
- ✅ Multi-critères (marge + fidélité + compétition)
- ✅ Transparent & observable

### vs. Rule-Based Systems
- ✅ Comportements émergents
- ✅ Adaptabilité dynamique
- ✅ Contexte client intégré

---

## Conclusion

### Démonstration Réussie
- Protocoles standards (UCP/MCP/A2A) fonctionnels
- Agents autonomes capables de décisions complexes
- Architecture extensible et observable

### Valeur Apportée
- **E-commerce** : pricing dynamique intelligent
- **Research** : plateforme expérimentation multi-agents
- **Education** : patterns UCP/MCP/A2A concrets

### Perspective
Les systèmes multi-agents ouvrent de nouvelles possibilités pour l'optimisation commerce... **avec prudence** !

---

## Questions ?

**Repository** : `github.com/owulveryck/ucp-merchant-test`

**Démos** :
- UCP Conformance : `go run ./sample_implementation`
- Shopping Demo : `demo/scripts/run_demo.sh`
- Arena : `demo/bin/arena`

**Documentation** :
- README.md
- QUICK_START.md
- docs/decisions/ (10 ADRs)

---

## Annexe - Quick Start Commands

```bash
# Build all
go build ./sample_implementation
go build -o demo/bin/arena ./demo/cmd/arena

# Run conformance server
go run ./sample_implementation --port 8182 \
  --data-dir test_data/flower_shop

# Run arena
demo/bin/arena
# → http://localhost:8888

# Run shopping demo
demo/scripts/run_demo.sh
```

---

## Annexe - Architecture Layers

```
┌─────────────────────────────────────────┐
│   Arena (5 Merchants Competition)      │
├─────────────────────────────────────────┤
│   Shopping Demo (Client + Graph)       │
├─────────────────────────────────────────┤
│   UCP Merchant Base (60 Tests ✅)       │
├─────────────────────────────────────────┤
│   Go Stdlib + HTTP + JSON-RPC          │
└─────────────────────────────────────────┘
```

Chaque couche indépendante et testable.

---

# Merci !

**Contact** : Stage OCTO
**Date** : Juin 2026
**Projet** : UCP Merchant Test - Pricing Intelligence
