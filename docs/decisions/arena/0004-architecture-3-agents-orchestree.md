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

Pour créer un système de pricing compétitif, il faut combiner deux analyses différentes :
- **Analyse Client** : adapter le prix selon le profil (fidélité, tier, historique)
- **Analyse Compétitive** : calculer un prix qui bat les concurrents

Comment structurer le système pour coordonner ces deux analyses ?

## Facteurs de Décision

* **Séparation des préoccupations** : Client vs Compétitivité = logiques différentes
* **Orchestration** : Besoin de coordonner les analyses
* **Transparence** : Montrer le raisonnement de chaque agent
* **Extensibilité** : Pouvoir ajouter d'autres agents facilement

## Options Considérées

* Option 1: Un seul agent qui fait tout
* Option 2: Deux agents indépendants
* Option 3: Trois agents avec orchestrateur central

## Décision

Option choisie: "**Option 3: Architecture 3-Agents orchestrée**" - un agent vendeur coordonne un agent client et un agent compétitivité.

### Conséquences

* Good, because séparation claire des responsabilités
* Good, because orchestration centralisée
* Good, because transparence (3 raisonnements visibles)
* Good, because extensible (ajouter Agent Inventaire, Publicité...)
* Bad, because plus complexe qu'un agent unique

### Implémentation

**Architecture** :

```
Agent 1 (Vendeur/Orchestrateur)
    ├── Agent 2 (Customer Growth)
    │   └── Analyse profil → Bonus -0% à -15%
    └── Agent 3 (Compétitivité)
        └── Shopping Graph → Prix optimal vs concurrents
```

**Agent 1 - Vendeur** : Coordonne et décide
**Agent 2 - Customer Growth** : Calcule bonus fidélité
**Agent 3 - Compétitivité** : Bat les concurrents

**Flux** :
```
1. Agent 2 : "Client Gold → bonus -10%"
2. Agent 3 : "Concurrent à $58 → prix $57"
3. Agent 1 : "$57 - 10% = $51.30" ✅
```

**Résultat** : Prix $51.30 gagne 100% des ventes

## Avantages et Inconvénients des Options

### Option 1: Agent Monolithique
* Good, because simple
* Bad, because logique mélangée
* Bad, because pas extensible

### Option 2: Deux Agents
* Good, because séparation
* Bad, because qui orchestre ?

### Option 3: Trois Agents (Choisi)
* Good, because orchestration claire
* Good, because extensible
* Bad, because plus de code

## Références

- Code : `pkg/pricing-unified/`
- Démo : `./scripts/arena_challenge.sh`
