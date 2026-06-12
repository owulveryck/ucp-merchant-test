---
status: accepté
date: 2026-05-29
---

# ADR-0001 : Architecture Multi-Agents pour Prix Compétitif

## Problème

Les marchands perdent des ventes sans comprendre pourquoi. Les concurrents affichent $60 mais vendent à $54 (code WELCOME10 caché). Les algorithmes monolithiques ne détectent pas ces réductions cachées.

## Décision

Architecture avec 4 agents spécialisés séquentiels :
- **Agent 1** : Détecte codes promo concurrents, calcule prix effectifs
- **Agent 2** : Analyse position marché
- **Agent 3** : Recommande stratégie pricing
- **Agent 4** : Valide contraintes marge/coût

## Pourquoi

- Transparent : Chaque agent explique son raisonnement
- Modulaire : Modifier un agent sans toucher les autres
- Extensible : Ajouter Agent 5 (pub) ou Agent 6 (stock) facilement

## Conséquences

**Positif**
- Marchands voient le raisonnement complet
- Modification indépendante de chaque agent
- Nouvelles capacités sans refonte

**Négatif**
- 4 fichiers au lieu d'1 fonction
- Latence séquentielle (<2s, acceptable)

## Validation

- Performance : <2s end-to-end
- Dashboard affiche raisonnement des 4 agents
- Tests unitaires par agent + tests d'intégration

## Implémentation

`pkg/merchant/competitive/orchestrator.go`
`pkg/merchant/competitive/agents/`
