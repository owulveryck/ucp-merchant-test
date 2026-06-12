---
parent: Decisions
nav_order: 5
title: ADR-005 Agent Acheteur Intégré

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer
---

# Agent Acheteur Intégré dans l'Interface Web

## Contexte et Problème

Pour tester le système de pricing, il faut un agent acheteur qui :
- Recherche des produits dans le Shopping Graph
- Compare les prix de tous les marchands
- Sélectionne le marchand le moins cher

Comment permettre à l'utilisateur de tester l'achat facilement depuis l'interface web ?

## Facteurs de Décision

* **Facilité d'utilisation** : Zero setup
* **Feedback temps réel** : Voir les étapes en direct
* **Intégration UX** : Interface fluide
* **Simplicité démo** : Facile à montrer

## Options Considérées

* Option 1: Agent externe (nécessite GCP, API keys)
* Option 2: Agent côté client JavaScript
* Option 3: Agent intégré côté serveur Go

## Décision

Option choisie: "**Option 3: Agent intégré serveur**" - fonction Go avec feedback temps réel via SSE.

### Conséquences

* Good, because zero setup
* Good, because feedback temps réel (SSE)
* Good, because facile à démontrer
* Good, because notifications visuelles (toast)
* Bad, because logique simple (pas de NLP)

### Implémentation

**Architecture** :

```
Interface Web
    ↓ POST /command
Serveur Go (executeBuyingFlow)
    ↓ SSE events
Interface (toast + messages)
```

**Flux utilisateur** :
```
1. Tape "Achète un casque" → Clique 🛒
2. Messages temps réel :
   🔍 Recherche...
   📊 Comparaison : MonMagasin $51.30 ← ✅
   🎯 DÉCISION : MonMagasin !
   ✅ Achat confirmé
3. Toast + marchand surligné
```

**Fichier** : `demo/internal/obs/handler.go` (executeBuyingFlow)

## Avantages et Inconvénients des Options

### Option 1: Agent Externe
* Bad, because setup complexe (GCP, API)

### Option 2: Client JavaScript  
* Bad, because pas d'accès Shopping Graph

### Option 3: Serveur Go (Choisi)
* Good, because zero setup
* Good, because SSE temps réel

## Références

- Code : `demo/internal/obs/handler.go`
- Interface : http://localhost:9002/arena
