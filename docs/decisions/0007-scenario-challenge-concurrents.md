---
parent: Decisions
nav_order: 7
title: ADR-007 Scénario Challenge

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer
---

# Scénario Challenge avec 4 Concurrents Pré-Créés pour Démo Rapide

## Contexte et Problème

Pour démontrer l'efficacité du système multi-agents, nous avons besoin d'un scénario reproductible qui montre clairement la valeur ajoutée.

Le script `demo.sh` existant nécessite :
- Création manuelle de 2-3 marchands (2-3 minutes)
- Configuration de leurs prix un par un
- Pas de scénario narratif clair
- Difficile de montrer un "avant/après" marquant

Pour une présentation au maître de stage ou une démo client, comment créer un scénario rapide (< 1 minute) qui montre de manière dramatique le passage de "perdant" à "gagnant" ?

## Facteurs de Décision

* **Rapidité** : Setup en moins d'1 minute
* **Reproductibilité** : Même résultat à chaque fois
* **Impact narratif** : Histoire claire "underdog → winner"
* **Simplicité** : Un seul script à lancer
* **Réalisme** : Simule une vraie situation de marché compétitif
* **Démonstration** : Facile à présenter lors d'une démo

## Options Considérées

* Option 1: Seulement demo.sh avec création manuelle
* Option 2: Auto-création de marchands au démarrage d'Arena
* Option 3: Script challenge séparé avec 4 concurrents pré-configurés

## Décision

Option choisie: "**Option 3: Script challenge séparé**", car il offre le meilleur compromis entre rapidité (30 secondes), impact narratif (scénario "David vs Goliath"), et flexibilité (demo.sh reste disponible pour exploration libre).

### Conséquences

* Good, because démo rapide (30 secondes vs 3 minutes)
* Good, because scénario narratif clair (4 concurrents établis vs nouveau marchand)
* Good, because reproductible (mêmes prix concurrents à chaque fois)
* Good, because impactant (passage de 0% à 100% de ventes)
* Good, because facile à présenter (un seul script)
* Good, because demo.sh reste disponible pour exploration
* Bad, because moins flexible (concurrents pré-configurés)
* Bad, because 2 scripts à maintenir (demo.sh + arena_challenge.sh)

### Confirmation

Le scénario est confirmé par :
- **Script** : `./scripts/arena_challenge.sh` fonctionne et crée 4 marchands
- **Timing** : Lancement + création marchands < 30 secondes
- **Résultat** : Marchand avec système gagne systématiquement vs les 4 concurrents
- **UX** : Instructions claires affichées étape par étape

### Implémentation

**Script `arena_challenge.sh`** :

```bash
#!/bin/bash

# 1. Lance les 3 services (shopping-graph, obs-hub, arena)

# 2. Crée 4 concurrents automatiquement
COMPETITORS=(
    "MegaStore:6200:mega_001"
    "PrixCassés:5800:prix_002"
    "SuperDeals:6000:super_003"
    "TopPrix:5900:top_004"
)

for merchant in "${COMPETITORS[@]}"; do
    # Inscription + configuration prix
    curl -X POST http://localhost:8888/register ...
    curl -X PUT http://localhost:8888/${id}/api/config ...
done

# 3. Affiche instructions scénario
echo "1. Crée ton marchand (sans optimiser)"
echo "2. Teste l'achat dans l'arène → concurrent gagne"
echo "3. Active système 3-agents → TU GAGNES !"
```

## Avantages et Inconvénients des Options

### Option 1: Seulement demo.sh (Création Manuelle)

Utilisateur crée tous les marchands manuellement.

* Good, because contrôle total sur le scénario
* Good, because flexibilité (prix, nombre de marchands)
* Good, because pédagogique (comprend chaque étape)
* Bad, because setup long (2-3 minutes)
* Bad, because pas reproductible (prix varient)
* Bad, because pas de scénario narratif clair
* Bad, because fastidieux pour une démo

### Option 2: Auto-Création au Démarrage

Arena crée automatiquement des marchands au démarrage.

* Good, because pas de script séparé
* Good, because automatique (zero action)
* Bad, because pas de contrôle (toujours les mêmes marchands)
* Bad, because pollue l'environnement (marchands permanents)
* Bad, because pas adapté pour exploration libre
* Bad, because moins de flexibilité

### Option 3: Script Challenge Séparé (Choisi)

Script dédié qui crée 4 concurrents pré-configurés.

* Good, because rapide (< 30 secondes)
* Good, because reproductible (mêmes concurrents)
* Good, because scénario narratif clair ("David vs Goliath")
* Good, because facile à démontrer (un script)
* Good, because demo.sh reste disponible
* Good, because concurrents réalistes (prix entre $58-$62)
* Neutral, because 2 scripts à maintenir
* Bad, because moins flexible que création manuelle

## Informations Complémentaires

### Scénario Narratif

**Acte 1 : Le Nouveau Marchand (Perdant)**
- 4 concurrents établis dominent le marché
- Prix compétitifs : $58, $59, $60, $62
- Nouveau marchand arrive avec prix manuel $70
- Résultat : 0% des ventes

**Acte 2 : L'Activation du Système**
- Marchand clique "💡 Calculer meilleur prix"
- Système 3-agents analyse en temps réel :
  - Agent 2 : Client Gold → bonus -10%
  - Agent 3 : Concurrent plus bas $58 → recommande $57
  - Agent 1 : Décision finale $51.30
- Prix appliqué

**Acte 3 : La Victoire (Gagnant)**
- Test d'achat dans l'arène
- Agent acheteur sélectionne le moins cher
- Résultat : 100% des ventes + notification toast
- Impact : Passage de 0% à 100% en un clic

### Configuration Concurrents

**4 marchands pré-créés avec prix réalistes** :

| Marchand | Prix de Base | Prix Final Système | Position |
|----------|--------------|-------------------|----------|
| PrixCassés | $58.00 | ~$57.50 | 1er (avant) |
| TopPrix | $59.00 | ~$58.50 | 2ème |
| SuperDeals | $60.00 | ~$59.50 | 3ème |
| MegaStore | $62.00 | ~$61.50 | 4ème |
| **MonMagasin** | **$70.00** | **$51.30** | **Dernier → 1er** ✨ |

### Impact Mesuré

**Sans système** :
- Prix : $70.00
- Écart vs concurrent : +20.7%
- Ventes : 0%

**Avec système** :
- Prix : $51.30
- Écart vs concurrent : -11.6%
- Ventes : 100%
- **ROI : +∞ (de 0 à 100%)**

### Instructions Affichées

Le script affiche 3 étapes claires :

```
🎯 DÉMO EN 3 ÉTAPES

1. Crée ton marchand (sans optimiser)
   → http://localhost:8888
   Laisse le prix par défaut (~$70) → tu seras le plus cher

2. Teste l'achat dans l'arène
   → http://localhost:9002/arena
   Tape "Achète un casque" → un concurrent gagne

3. Active le système 3-agents
   Dashboard → "💡 Calculer meilleur prix" → Applique
   Retour arène → "Achète un casque" → TU GAGNES !
```

### Timing Démo

- **0:00-0:30** : Lancement script + création 4 concurrents
- **0:30-1:00** : Création de ton marchand, prix $70
- **1:00-1:30** : Test achat → concurrent gagne
- **1:30-2:00** : Activation système 3-agents
- **2:00-2:30** : Test achat → TU GAGNES

**Total : 2min30** pour démo complète vs 5-6 minutes avec demo.sh

### Comparaison Scripts

| Critère | demo.sh | arena_challenge.sh |
|---------|---------|-------------------|
| **Setup** | Création manuelle | 4 concurrents auto |
| **Temps** | 3-5 minutes | 30 secondes |
| **Flexibilité** | Totale | Limitée |
| **Scénario** | Exploration | Challenge |
| **Reproductibilité** | Faible | Haute |
| **Impact démo** | Moyen | Fort |
| **Usage** | Développement | Présentation |

### Références

- Script : `scripts/arena_challenge.sh`
- Commit `3723498` (2026-06-04) : feat: Système multi-agents 3-agents
- Lié à : ADR-004 (Architecture 3-Agents), ADR-005 (Agent Acheteur Intégré)
