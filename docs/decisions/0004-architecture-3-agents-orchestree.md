---
parent: Decisions
nav_order: 4
title: ADR-004 Architecture 3-Agents Orchestrée

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer, Olivier Wulveryck
---

# Architecture 3-Agents pour Pricing Compétitif Intelligent

## Contexte et Problème

Pour créer un système de pricing compétitif, il faut résoudre deux problèmes distincts :

1. **Analyse Client** : Adapter les prix selon le profil client (fidélité, tier, historique d'achats)
2. **Analyse Compétitive** : Calculer un prix qui bat les concurrents tout en préservant la marge

Ces deux analyses nécessitent des données et une logique différentes :
- L'analyse client utilise le profil utilisateur, l'historique, les bonus fidélité
- L'analyse compétitive interroge le Shopping Graph pour connaître les prix concurrents

De plus, il faut une **orchestration** pour coordonner ces deux analyses et prendre une décision finale cohérente.

Comment structurer le système pour séparer ces préoccupations tout en permettant une décision coordonnée ?

## Facteurs de Décision

* **Séparation des préoccupations** : Client vs Compétitivité = logiques différentes
* **Orchestration** : Besoin de coordonner les 2 analyses
* **Transparence** : Chaque agent doit expliquer son raisonnement
* **Extensibilité** : Pouvoir ajouter d'autres agents (Inventaire, Publicité)
* **Maintenabilité** : Code clair et modulaire

## Options Considérées

* Option 1: Un seul agent monolithique
* Option 2: Deux agents indépendants (Client + Compétitivité)
* Option 3: Trois agents avec orchestration (Vendeur + Client + Compétitivité)

## Décision

Option choisie: "**Option 3: Architecture 3-Agents avec orchestration**", car elle sépare clairement les responsabilités (client vs compétitivité) tout en ayant un orchestrateur central (Vendeur) qui coordonne et prend la décision finale.

### Conséquences

* Good, because séparation claire : analyse client ≠ analyse compétitive
* Good, because orchestration centralisée (Agent 1 coordonne)
* Good, because extensible (facile d'ajouter Agent 4 Inventaire, Agent 5 Publicité)
* Good, because transparence totale (raisonnement de chaque agent visible)
* Good, because chaque agent est testable indépendamment
* Bad, because plus complexe qu'un agent unique
* Bad, because coordination entre agents nécessaire

### Confirmation

L'architecture est confirmée par :
- **Tests réels** : Marchands avec système 3-agents gagnent systématiquement vs concurrents
- **Code** : `pkg/pricing-unified/` implémente les 3 agents
- **Dashboard** : Affichage du raisonnement des 3 agents en temps réel
- **Démo** : Scénario challenge montre marchand passant de perdant (prix manuel $70) à gagnant (système calcule $51.30)

### Implémentation

Architecture en 3 agents spécialisés :

```
┌─────────────────────────────────────────────────────────┐
│           Agent 1 : Vendeur (Orchestrateur)             │
│   Coordonne Agent 2 et 3, décision finale               │
└────────────────┬────────────────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
┌──────────────┐  ┌──────────────────────────────────────┐
│  Agent 2:    │  │  Agent 3: Compétitivité              │
│  Customer    │  │                                       │
│  Growth      │  │  • Interroge Shopping Graph          │
│              │  │  • Compare prix concurrents          │
│  • Profil    │  │  • Détecte codes promo               │
│  • Tier      │  │  • Recommande prix optimal           │
│  • Bonus     │  │  • Valide marge minimum              │
│  • -0% à     │  │                                       │
│    -15%      │  │                                       │
└──────────────┘  └──────────────────────────────────────┘
```

**Agent 1 - Vendeur (Orchestrateur)** :
- Coordonne Agent 2 et Agent 3
- Synthétise leurs recommandations
- Prend la décision finale de pricing
- Fichier : `pkg/pricing-unified/orchestrator.go`

**Agent 2 - Customer Growth** :
- Analyse le profil client (nouveau, régulier, VIP)
- Détecte le tier (Bronze, Silver, Gold)
- Calcule bonus fidélité (0% à -15%)
- Fichier : `pkg/pricing-unified/agents/customer_growth.go`

**Agent 3 - Compétitivité** :
- Interroge Shopping Graph pour prix concurrents
- Détecte codes promo concurrents
- Calcule prix pour battre la concurrence
- Valide que la marge minimum est respectée
- Fichier : `pkg/pricing-unified/agents/competitiveness.go`

## Avantages et Inconvénients des Options

### Option 1: Agent Monolithique

Un seul agent qui fait tout (client + compétitivité + décision).

* Good, because simple (un seul agent)
* Good, because pas de coordination nécessaire
* Good, because moins de code
* Bad, because logique mélangée (client + compétitivité)
* Bad, because difficile à tester (tout couplé)
* Bad, because pas extensible (comment ajouter inventaire ?)
* Bad, because pas transparent (un seul raisonnement global)

### Option 2: Deux Agents Indépendants

Agent Client + Agent Compétitivité sans orchestrateur.

* Good, because séparation client vs compétitivité
* Good, because chaque agent est testable
* Neutral, because besoin de combiner les 2 résultats
* Bad, because qui décide entre les 2 recommandations ?
* Bad, because pas d'orchestration (logique de combinaison dispersée)
* Bad, because difficile d'ajouter un 3ème agent

### Option 3: Trois Agents avec Orchestration (Choisi)

Agent Vendeur orchestre Agent Client + Agent Compétitivité.

* Good, because séparation claire des responsabilités
* Good, because orchestration centralisée (Agent 1)
* Good, because extensible (ajouter Agent 4, 5... facilement)
* Good, because transparence (3 raisonnements visibles)
* Good, because testable indépendamment
* Neutral, because architecture en couches
* Bad, because plus complexe qu'un agent unique

## Informations Complémentaires

### Flux de Décision

```
1. Utilisateur clique "💡 Calculer meilleur prix"
   ↓
2. Agent 1 (Vendeur) lance orchestration
   ↓
3. Agent 2 (Customer Growth)
   → Détecte client Gold
   → Recommande bonus -10%
   → Retourne : "Bonus fidélité : -10%"
   ↓
4. Agent 3 (Compétitivité)
   → Interroge Shopping Graph
   → Trouve concurrent le plus bas : $58.00
   → Calcule prix pour gagner : $57.00
   → Valide marge (> coût)
   → Retourne : "Prix compétitif : $57.00"
   ↓
5. Agent 1 (Vendeur) synthétise
   → Prix de base : $57.00 (Agent 3)
   → Bonus client : -10% (Agent 2)
   → Décision finale : $57.00 × 0.9 = $51.30
   ↓
6. Dashboard affiche raisonnement des 3 agents
```

### Exemple Concret

**Situation** : 4 concurrents à $58, $59, $60, $62. Client Gold. Coût produit $45.

**Agent 2 - Customer Growth** :
```
🧑 AGENT 2 : Customer Growth

Analyse du profil client :
• Tier : Gold (15+ achats)
• Fidélité : Très élevée
• Historique : $850 dépensés

Recommandation :
• Bonus fidélité : -10%
• Rationale : Récompenser client VIP, augmenter rétention
```

**Agent 3 - Compétitivité** :
```
📊 AGENT 3 : Compétitivité

Analyse du marché (Shopping Graph) :
• 4 concurrents trouvés
• Prix le plus bas : $58.00 (PrixCassés)
• Notre coût : $45.00

Recommandation :
• Prix optimal : $57.00
• Rationale : Battre concurrent le plus bas de $1
• Marge : $57 - $45 = $12 (26.7%) ✅ OK
```

**Agent 1 - Vendeur (Décision finale)** :
```
🎯 AGENT 1 : Vendeur

Synthèse :
• Prix compétitif (Agent 3) : $57.00
• Bonus client (Agent 2) : -10%
• Calcul : $57.00 × 0.9 = $51.30

DÉCISION FINALE : $51.30

Rationale :
• Bat tous les concurrents (-11.6% vs meilleur)
• Récompense client Gold
• Marge finale : $51.30 - $45 = $6.30 (14%) ✅
```

**Résultat** : Prix $51.30 → Gagne 100% des ventes

### Impact Mesurable

**Démo arena_challenge.sh** :
- Avant système : 0% de ventes (prix $70 trop élevé)
- Après système : 100% de ventes (prix $51.30 optimal)
- Économie client vs concurrent : 11.6%

### Extensibilité

L'architecture permet d'ajouter facilement de nouveaux agents :

**Agent 4 - Inventaire** (futur) :
- Recommande prix plus élevé si stock faible
- Recommande prix plus bas si stock excessif

**Agent 5 - Publicité** (futur) :
- Ajuste prix selon campagnes en cours
- Prix attractif si pub active pour attirer trafic

L'orchestrateur (Agent 1) coordonne tous les agents :
```go
recommendations := []Recommendation{
    agent2.AnalyzeCustomer(ctx),
    agent3.AnalyzeCompetition(ctx),
    agent4.AnalyzeInventory(ctx),   // futur
    agent5.AnalyzeCampaigns(ctx),   // futur
}
finalPrice := orchestrator.Decide(recommendations)
```

### Code Clé

**Fichiers** :
- `pkg/pricing-unified/orchestrator.go` - Agent 1 (Vendeur)
- `pkg/pricing-unified/agents/customer_growth.go` - Agent 2
- `pkg/pricing-unified/agents/competitiveness.go` - Agent 3
- `demo/cmd/arena/tenant_3agents.go` - Configuration Arena

### Références

- Commit `3723498` (2026-06-04) : feat: Système multi-agents 3-agents
- Interface : http://localhost:8888 (dashboard marchand)
- Démo : `./scripts/arena_challenge.sh`
- Lié à : ADR-001 (Architecture Multi-Agents de base)
