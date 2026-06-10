# Tutorial : Comprendre le Pricing Multi-Agents

**Durée estimée** : 30 minutes  
**Prérequis** : Tutorials 1 & 2 terminés

## Objectif

À la fin de ce tutorial, vous aurez :
- ✅ Lancé l'Arena avec 5 marchands en compétition
- ✅ Compris le système 3-agents (Vendor, Customer Growth, Competitiveness)
- ✅ Observé une guerre des prix en temps réel
- ✅ Analysé les décisions de pricing via le dashboard

## Le système 3-agents expliqué

```
Agent 1 : Vendor Orchestrator (Le Chef d'orchestre)
    ↓
    ├─→ Agent 2 : Customer Growth (Le Fidélisateur)
    │   └─→ "Ce client est-il VIP ? Faut-il le garder ?"
    │   └─→ Output : shouldRetain, tier (gold/silver/bronze), discount%
    │
    └─→ Agent 3 : Competitiveness (Le Stratège Prix)
        └─→ "Quelle est notre position face aux concurrents ?"
        └─→ Wraps 4-agent system :
            ├─→ Price Intelligence (analyse concurrence)
            ├─→ Market Analysis (tendances)
            ├─→ Strategy Selection (balanced/premium/aggressive)
            └─→ Margin Validation (vérifie rentabilité)
```

## Étape 1 : Lancer l'Arena

```bash
cd ~/stageocto/ucp-merchant-test

# Compiler l'arena (si pas déjà fait)
go build -o demo/bin/arena ./demo/cmd/arena

# Lancer
demo/bin/arena
```

**Résultat attendu** :
```
Arena starting on http://localhost:8888
Landing:   http://localhost:8888/
Merchants: MegaStore, PrixCassés, SuperDeals, TopPrix, MonMarchand
```

## Étape 2 : Ouvrir l'interface web

Ouvrez http://localhost:8888 dans votre navigateur.

Vous verrez :
- **Liste des 5 marchands** avec leurs positions
- **Produit actuel** (ex: casque_audio)
- **Pricing en temps réel** avec décisions agents
- **Timeline des events** SSE

## Étape 3 : Observer une décision de pricing

Dans l'interface, cliquez sur **"Test Pricing"** pour un marchand.

**Ce qui se passe en coulisses** :

1. **Agent 1 reçoit la demande** : "Prix pour casque_audio, client default_customer"

2. **Agent 2 analyse le client** :
   ```
   [Customer Growth] Analyzing customer: default_customer
   Decision: ShouldRetain=true, Tier=gold, Discount=10%
   ```

3. **Agent 3 analyse la concurrence** :
   ```
   [Competitiveness] Analyzing product: casque_audio at $62.15
   [Price Intelligence] Got 4 competitors
   Competitor MegaStore: $61.17
   Competitor SuperDeals: $63.31
   Position: 2/5 (follower)
   Strategy: balanced, target: $58.11
   ```

4. **Agent 1 synthétise** :
   ```
   Agent 2: Keep client (gold) → -10% discount
   Agent 3: Competitive price → $58.11
   Final price: $52.30 (marge 4%)
   ```

## Étape 4 : Provoquer une guerre des prix

1. **Notez le prix actuel** de MonMarchand (ex: $62.15)
2. **Cliquez "Test Pricing"** → nouveau prix plus bas (ex: $52.30)
3. **Le prix est publié** dans le Shopping Graph
4. **Les autres marchands voient** ce nouveau prix bas
5. **Cliquez "Test Pricing"** à nouveau → encore plus bas ! (ex: $46.13)

**Résultat attendu** : Les prix descendent en cascade, parfois en marge négative !

## Étape 5 : Analyser les logs détaillés

Dans le terminal où tourne l'arena, observez :

```
orchestrator.go:169: Price Intelligence: rank 2/5, lowest: $61.17
orchestrator.go:189: Market Analysis: follower position, stable trend
orchestrator.go:210: Strategy: balanced, target: $58.11, confidence: 80%
orchestrator.go:239: ✅ Pricing approved: $58.11 (discount: $4.04, margin: 13%)
orchestrator.go:139: ✓ Prix final décidé: $52.30 (marge 4%)
```

Puis au 2ème tour :

```
orchestrator.go:169: Price Intelligence: rank 1/5, lowest: $52.30 (ourselves!)
orchestrator.go:189: Market Analysis: leader position, down trend
orchestrator.go:210: Strategy: premium, target: $51.25
orchestrator.go:235: ⚠️ Pricing adjusted: [⚠️ Marge réduite: 2% pour GAGNER]
orchestrator.go:139: ✓ Prix final décidé: $46.13 (marge -8%!)
```

## Étape 6 : Observer les fichiers de logs

```bash
# Voir les logs de l'arena
tail -f logs/arena.log

# Voir les logs du Shopping Graph
tail -f logs/shopping-graph.log
```

## Ce que vous avez appris

✅ **Architecture 3-agents** : orchestration, spécialisation  
✅ **Agent 2** : fidélisation client (tier, discount)  
✅ **Agent 3** : intelligence compétitive (position marché, stratégie)  
✅ **Trade-off marge vs victoire** : peut sacrifier rentabilité pour gagner  
✅ **Events SSE** : monitoring temps réel des décisions  

## Les 4 stratégies de pricing

| Stratégie | Quand | Objectif | Marge typique |
|-----------|-------|----------|---------------|
| `premium` | Leader | Maximiser marge | 15-20% |
| `balanced` | Milieu de marché | Équilibre | 10-15% |
| `aggressive` | Follower | Gagner parts | 5-10% |
| `vip_retention` | Client premium | Fidéliser | Variable |

## Les 4 agents sous Agent 3

```
Agent 3 (Competitiveness) wraps:
├─→ Agent 3.1 : Price Intelligence
│   └─→ Récupère prix concurrents via Shopping Graph
│   └─→ Calcule position (rank X/N)
│
├─→ Agent 3.2 : Market Analysis
│   └─→ Analyse tendance (up/down/stable)
│   └─→ Identifie opportunités (match_market/optimize/defend)
│
├─→ Agent 3.3 : Strategy Selection
│   └─→ Choisit stratégie (premium/balanced/aggressive)
│   └─→ Calcule target price
│
└─→ Agent 3.4 : Margin Validation
    └─→ Vérifie marge minimale (10% par défaut)
    └─→ Peut accepter moins si nécessaire pour gagner
```

## Exercice pratique

**Défi** : Faire que MonMarchand garde une marge >= 10%

**Solution** : Modifier `pkg/pricing-unified/agents/orchestrator.go` ligne ~235 :

```go
// Actuel : accepte marge réduite
if finalMargin < minMarginPercent {
    warnings = append(warnings, fmt.Sprintf("⚠️ Marge réduite: %d%% pour GAGNER", finalMargin))
}

// Modifier en :
if finalMargin < minMarginPercent {
    return PriceDecision{}, fmt.Errorf("marge insuffisante: %d%% < %d%%", finalMargin, minMarginPercent)
}
```

Recompiler et observer : MonMarchand refuse les prix non-rentables !

## Prochaines étapes

- **How-to** : [Configurer des stratégies de pricing personnalisées](../how-to/setup-pricing-strategies.md)
- **Reference** : [Architecture détaillée des agents](../reference/agent-architecture.md)
- **Explanation** : [Pourquoi multi-agents ?](../explanation/why-multi-agent.md)

## En cas de problème

**Les marchands ne se voient pas** :
- Vérifiez que le Shopping Graph tourne
- Vérifiez les logs : `tail -f logs/shopping-graph.log`

**Les prix ne changent pas** :
- Cliquez "Refresh" pour recharger les données
- Vérifiez les logs agents dans `logs/arena.log`

**Marge toujours négative** :
- C'est normal avec la stratégie "victoire avant tout" !
- Voir [Explanation: Trade-offs compétitifs](../explanation/competitive-tradeoffs.md)
