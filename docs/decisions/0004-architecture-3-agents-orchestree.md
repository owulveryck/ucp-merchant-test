---
parent: Decisions
nav_order: 4
title: ADR-004 Architecture 3-Agents Orchestrée

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer, Olivier Wulveryck
---

# Architecture 3-Agents Enveloppant le Système 4-Agents Existant

## Contexte et Problème

Le système 4-agents de pricing compétitif fonctionne mais présente des limitations :
- Pas d'orchestration centralisée entre les agents
- Pas de prise en compte du profil client (fidélité, tier, historique)
- Pas de coordination entre analyse client et analyse compétitive
- Décisions prises sans vue d'ensemble sur la stratégie commerciale

Comment ajouter une couche d'orchestration et d'analyse client tout en préservant la logique compétitive éprouvée du système 4-agents ?

## Facteurs de Décision

* **Préservation de l'existant** : Le système 4-agents fonctionne et ne doit pas être perdu
* **Analyse client** : Besoin d'une couche pour gérer fidélité, tiers client, bonus
* **Orchestration** : Besoin de coordonner analyse client et analyse compétitive
* **Transparence** : Chaque agent doit expliquer son raisonnement
* **Extensibilité** : Pouvoir ajouter d'autres agents facilement
* **Maintenabilité** : Code clair et modulaire

## Options Considérées

* Option 1: Remplacer le système 4-agents par un nouveau système
* Option 2: Ajouter les nouveaux agents en parallèle du système 4-agents
* Option 3: Architecture hybride - 3 agents dont un enveloppe le système 4-agents

## Décision

Option choisie: "**Option 3: Architecture Hybride 3-Agents**", car elle préserve la logique éprouvée tout en ajoutant orchestration et analyse client. L'Agent 3 (Compétitivité) enveloppe les 4 agents existants, réutilisant leur expertise sans duplication.

### Conséquences

* Good, because réutilise logique 4-agents éprouvée (détection promo, analyse marché, stratégie, validation)
* Good, because ajoute couche client manquante (profil, fidélité, bonus)
* Good, because orchestration centralisée par Agent 1 (Vendeur)
* Good, because extensible (facile d'ajouter Agent 4 Inventaire, Agent 5 Publicité)
* Good, because transparence totale (raisonnement de chaque agent visible)
* Bad, because complexité (7 agents au total : 3 nouveaux + 4 existants)
* Bad, because courbe d'apprentissage (comprendre les 2 niveaux)

### Confirmation

L'architecture est confirmée par :
- **Tests réels** : Marchands avec système 3-agents gagnent systématiquement vs concurrents
- **Code** : `pkg/pricing-unified/` implémente les 3 agents, Agent 3 délègue à `competitive.Orchestrator`
- **Dashboard** : Affichage du raisonnement des 3 agents en temps réel
- **Démo** : Scénario challenge montre marchand passant de perdant (prix manuel $70) à gagnant (système calcule $51.30)

### Implémentation

Architecture en 3 couches :

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
│  Customer    │  │  ┌────────────────────────────────┐  │
│  Growth      │  │  │  Système 4-Agents Existant     │  │
│              │  │  │  • Agent 1: Détection promo    │  │
│  • Profil    │  │  │  • Agent 2: Analyse marché     │  │
│  • Tier      │  │  │  • Agent 3: Stratégie          │  │
│  • Bonus     │  │  │  • Agent 4: Validation         │  │
│  • -0% à     │  │  └────────────────────────────────┘  │
│    -15%      │  │  Interroge Shopping Graph             │
└──────────────┘  └──────────────────────────────────────┘
```

**Fichiers clés** :
- `pkg/pricing-unified/orchestrator.go` - Agent 1 (Vendeur)
- `pkg/pricing-unified/agents/customer_growth.go` - Agent 2
- `pkg/pricing-unified/agents/competitiveness.go` - Agent 3 (enveloppe)
- `demo/cmd/arena/tenant_3agents.go` - Configuration Arena

## Avantages et Inconvénients des Options

### Option 1: Remplacer le Système 4-Agents

Réécrire complètement le système de pricing.

* Good, because architecture plus simple (seulement 3 agents)
* Good, because pas de "legacy code"
* Bad, because perte de logique éprouvée (détection promo codes, stratégies)
* Bad, because risque de régresser en qualité de décisions
* Bad, because temps de développement élevé (tout recoder)

### Option 2: Agents en Parallèle

Ajouter 2 nouveaux agents qui coexistent avec les 4 existants.

* Good, because garde système 4-agents intact
* Good, because simple à implémenter
* Bad, because pas d'orchestration (qui décide entre les 6 agents ?)
* Bad, because pas de coordination client + compétitivité
* Bad, because confusion sur qui a l'autorité finale

### Option 3: Architecture Hybride (Choisi)

3 agents dont Agent 3 enveloppe le système 4-agents.

* Good, because réutilise expertise 4-agents existants
* Good, because orchestration claire (Agent 1 coordonne)
* Good, because séparation des préoccupations (client vs compétitivité)
* Good, because extensible (ajouter Agent 4 Inventaire facilement)
* Good, because transparence (dashboard montre les 3 niveaux)
* Neutral, because architecture en couches (peut sembler complexe)
* Bad, because 7 agents au total à comprendre
* Bad, because abstraction supplémentaire (Agent 3 = wrapper)

## Informations Complémentaires

### Flux de Décision

```
1. Utilisateur clique "Calculer meilleur prix"
   ↓
2. Agent 1 (Vendeur) lance orchestration
   ↓
3. Agent 2 (Customer Growth)
   → Détecte client Gold
   → Recommande bonus -10%
   ↓
4. Agent 3 (Compétitivité)
   → Enveloppe système 4-agents :
     • Agent sous-jacent 1 : Détecte codes promo concurrents
     • Agent sous-jacent 2 : Analyse position marché
     • Agent sous-jacent 3 : Stratégie pricing (battre $58 → $57)
     • Agent sous-jacent 4 : Validation marge (OK)
   → Recommande $57.00
   ↓
5. Agent 1 (Vendeur) synthétise
   → Prix compétitif : $57.00
   → Bonus client : -10%
   → Décision finale : $51.30
   ↓
6. Dashboard affiche raisonnement des 3 agents
```

### Exemple Concret

**Situation** : 4 concurrents à $58, $59, $60, $62. Marchand arrive avec prix manuel $70.

**Sans système** : Prix $70 → Perd toutes les ventes

**Avec système 3-agents** :
- Agent 2 : "Client Gold détecté, bonus -10%"
- Agent 3 : "Concurrent le plus bas $58, recommande $57 pour gagner"
- Agent 1 : "Prix final = $57 - 10% = $51.30"

**Résultat** : Prix $51.30 → Gagne 100% des ventes

### Impact Mesurable

**Démo arena_challenge.sh** :
- Avant système : 0% de ventes (prix $70 trop élevé)
- Après système : 100% de ventes (prix $51.30 optimal)
- Économie client vs 2ème meilleur : 11.6%

### Références

- Commit `3723498` (2026-06-04) : feat: Système multi-agents 3-agents
- Fichiers : `pkg/pricing-unified/`, `demo/cmd/arena/tenant_3agents.go`
- Lié à : ADR-001 (Architecture Multi-Agents pour Pricing Compétitif)
