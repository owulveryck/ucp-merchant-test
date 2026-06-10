---
parent: Decisions
nav_order: 6
title: ADR-006 Messages Détaillés Décision d'Achat

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer
---

# Messages Détaillés de Décision d'Achat pour Transparence et Validation

## Contexte et Problème

Lorsque l'agent acheteur intégré sélectionne un marchand dans l'arène, l'utilisateur voit seulement le résultat final ("MonMagasin sélectionné") sans comprendre :
- Pourquoi ce marchand a été choisi
- Quels autres marchands étaient disponibles
- Quel était l'écart de prix avec les concurrents
- Si la décision est correcte (le moins cher)

Cette opacité pose des problèmes :
- Difficile de valider que le système fonctionne correctement
- Pas de confiance dans la décision de l'agent
- Impossible de déboguer si quelque chose ne va pas
- Démo moins convaincante (résultat sans justification)

Comment rendre le processus de décision de l'agent acheteur transparent et vérifiable ?

## Facteurs de Décision

* **Transparence** : L'utilisateur doit comprendre pourquoi un marchand est sélectionné
* **Validation** : Facile de vérifier que c'est bien le moins cher
* **Débogage** : En cas d'erreur, identifier rapidement le problème
* **Pédagogie** : Comprendre le raisonnement de l'agent
* **Confiance** : Prouver que le système fonctionne comme prévu
* **Démo** : Présentation convaincante avec justification claire

## Options Considérées

* Option 1: Message simple "X gagne"
* Option 2: Afficher seulement le prix final
* Option 3: Comparaison détaillée + décision justifiée

## Décision

Option choisie: "**Option 3: Comparaison détaillée + décision justifiée**", car elle offre transparence totale, validation facile, et rend la démo beaucoup plus convaincante en montrant le raisonnement complet de l'agent.

### Conséquences

* Good, because transparence totale du processus de décision
* Good, because validation immédiate (voir tous les prix comparés)
* Good, because débogage facile (identifier rapidement les erreurs)
* Good, because démo plus convaincante (preuve chiffrée)
* Good, because pédagogique (comprendre comment l'agent raisonne)
* Good, because confiance renforcée (preuve que c'est le moins cher)
* Bad, because plus de messages affichés (peut être verbeux)
* Bad, because plus de code pour formater les messages
* Neutral, because utilisateurs peuvent fermer le panel s'ils ne veulent pas voir les détails

### Confirmation

L'implémentation est confirmée par :
- **Code** : `demo/internal/obs/handler.go` lignes 349-410 (`executeBuyingFlow`)
- **Interface** : Panel d'activité dans http://localhost:9002/arena affiche les 3 types de messages
- **Tests** : Scénario challenge montre comparaison avec 4 concurrents
- **UX** : Messages structurés avec émojis pour faciliter la lecture

### Implémentation

**3 types de messages envoyés par l'agent** :

```go
// 1. Message de comparaison des prix
message := "📊 Comparaison des prix :\n"
for _, merchant := range allMerchants {
    if merchant.ID == cheapest.ID {
        message += fmt.Sprintf("   • %s: $%.2f ← ✅ LE MOINS CHER\n", 
            merchant.Name, merchant.Price)
    } else {
        diff := merchant.Price - cheapest.Price
        message += fmt.Sprintf("   • %s: $%.2f (+$%.2f)\n", 
            merchant.Name, merchant.Price, diff)
    }
}

// 2. Message de décision avec justification
decision := fmt.Sprintf(`🎯 DÉCISION : %s sélectionné !

Pourquoi ?
   • Prix le plus bas : $%.2f
   • %d concurrent(s) comparé(s)
   • Économie : %.1f%% vs 2ème meilleur prix`,
    cheapest.Name, cheapest.Price, 
    len(allMerchants)-1, savings)

// 3. Message de confirmation d'achat
confirmation := fmt.Sprintf("✅ Achat confirmé ! Prix final: $%.2f (avec AUTO_COMPETE)", 
    cheapest.Price)
```

**Affichage dans l'interface** :

```
🤖 Agent : 🔍 Recherche: Achète un casque

🤖 Agent : 📊 Comparaison des prix :
   • MonMagasin: $51.30 ← ✅ LE MOINS CHER
   • PrixCassés: $58.00 (+$6.70)
   • TopPrix: $59.00 (+$7.70)
   • SuperDeals: $60.00 (+$8.70)
   • MegaStore: $62.00 (+$10.70)

🤖 Agent : 🎯 DÉCISION : MonMagasin sélectionné !

Pourquoi ?
   • Prix le plus bas : $51.30
   • 4 concurrent(s) comparé(s)
   • Économie : 11.6% vs 2ème meilleur prix

🤖 Agent : 🛒 Création du panier...

🤖 Agent : ✅ Achat confirmé ! Prix final: $51.30 (avec AUTO_COMPETE)
```

## Avantages et Inconvénients des Options

### Option 1: Message Simple "X Gagne"

Un seul message indiquant le marchand sélectionné.

* Good, because très concis (1 ligne)
* Good, because simple à implémenter
* Good, because pas de surcharge d'information
* Bad, because pas de transparence (pourquoi X ?)
* Bad, because impossible de valider la décision
* Bad, because pas convaincant en démo (résultat sans preuve)
* Bad, because difficile à déboguer si erreur
* Bad, because pas pédagogique

### Option 2: Afficher Seulement le Prix Final

Message avec le prix du marchand sélectionné uniquement.

* Good, because simple et court
* Good, because donne l'info essentielle (le prix)
* Neutral, because un peu plus informatif que juste le nom
* Bad, because pas de comparaison (est-ce vraiment le moins cher ?)
* Bad, because pas de justification
* Bad, because toujours opaque (pourquoi ce marchand ?)
* Bad, because difficile de valider sans voir les autres prix

### Option 3: Comparaison Détaillée + Décision Justifiée (Choisi)

Trois messages : comparaison complète, décision justifiée, confirmation.

* Good, because transparence totale (tous les prix visibles)
* Good, because validation immédiate (voir que c'est bien le moins cher)
* Good, because justification chiffrée (économie en %)
* Good, because facile à déboguer (voir si un marchand manque)
* Good, because démo convaincante (preuve chiffrée du gain)
* Good, because pédagogique (comprendre le raisonnement)
* Good, because émojis pour faciliter la lecture (📊, 🎯, ✅)
* Neutral, because plus de texte (mais panel scrollable)
* Bad, because plus de code pour formater
* Bad, because peut sembler verbeux si beaucoup de marchands (> 10)

## Informations Complémentaires

### Structure des Messages

**Message 1 : Comparaison des Prix** (📊)
- Liste TOUS les marchands avec leurs prix
- Marque clairement le moins cher avec ✅
- Affiche l'écart de prix (+$X.XX) pour chaque concurrent
- Ordre : gagnant en premier, puis concurrents triés par prix

**Message 2 : Décision Justifiée** (🎯)
- Annonce le marchand sélectionné
- Section "Pourquoi ?" avec 3 métriques clés :
  1. Prix le plus bas (valeur absolue)
  2. Nombre de concurrents comparés (crédibilité)
  3. Économie vs 2ème meilleur prix (% de gain)

**Message 3 : Confirmation** (✅)
- Confirme la création du panier
- Affiche le prix final
- Mentionne l'utilisation du code AUTO_COMPETE

### Calcul de l'Économie

```go
// Trouve le 2ème meilleur prix pour calculer l'économie
secondBest := math.MaxFloat64
for _, m := range allMerchants {
    if m.ID != cheapest.ID && m.Price < secondBest {
        secondBest = m.Price
    }
}

// Calcule le % d'économie
savings := ((secondBest - cheapest.Price) / secondBest) * 100
```

### Exemple Concret : Scénario Challenge

**Contexte** : 5 marchands proposent un casque
- MonMagasin : $51.30 (avec système 3-agents activé)
- PrixCassés : $58.00
- TopPrix : $59.00
- SuperDeals : $60.00
- MegaStore : $62.00

**Affichage** :

```
📊 Comparaison des prix :
   • MonMagasin: $51.30 ← ✅ LE MOINS CHER
   • PrixCassés: $58.00 (+$6.70)
   • TopPrix: $59.00 (+$7.70)
   • SuperDeals: $60.00 (+$8.70)
   • MegaStore: $62.00 (+$10.70)
```

→ **Validation immédiate** : On voit que MonMagasin est effectivement le moins cher de $6.70

```
🎯 DÉCISION : MonMagasin sélectionné !

Pourquoi ?
   • Prix le plus bas : $51.30
   • 4 concurrent(s) comparé(s)
   • Économie : 11.6% vs 2ème meilleur prix
```

→ **Justification chiffrée** : 11.6% d'économie = ($58.00 - $51.30) / $58.00

```
✅ Achat confirmé ! Prix final: $51.30 (avec AUTO_COMPETE)
```

→ **Confirmation** : Transaction réussie avec le meilleur prix

### Impact sur la Démo

**Sans messages détaillés** :
```
🤖 Agent : MonMagasin sélectionné
```
→ Réaction : "OK... pourquoi ? Est-ce vraiment le moins cher ?"

**Avec messages détaillés** :
```
📊 Comparaison : 5 marchands, MonMagasin $51.30 vs autres $58-62
🎯 Décision : 11.6% d'économie, gagne vs 4 concurrents
✅ Confirmé : $51.30
```
→ Réaction : "Wow, preuve claire que le système fonctionne ! 11.6% de gain !"

### Format pour Lisibilité

**Émojis utilisés** :
- 🔍 : Recherche initiale
- 📊 : Comparaison de données
- 🎯 : Décision prise
- 🛒 : Action de création panier
- ✅ : Confirmation succès
- ← : Indicateur visuel "choisi"

**Indentation** :
- 3 espaces pour les listes de marchands
- Alignement des prix ($XX.XX)
- Séparation claire entre sections

**Structure** :
- Titre en gras (émoji + texte)
- Liste à puces pour données multiples
- Section "Pourquoi ?" pour justification

### Cas Limites

**1 seul marchand disponible** :
```
📊 Comparaison des prix :
   • MonMagasin: $51.30 ← ✅ SEUL DISPONIBLE

🎯 DÉCISION : MonMagasin sélectionné !
Pourquoi ?
   • Seul marchand avec le produit en stock
   • Prix : $51.30
```

**Égalité de prix** (2 marchands au même prix) :
```
📊 Comparaison des prix :
   • MonMagasin: $58.00 ← ✅ LE MOINS CHER (ex-aequo)
   • PrixCassés: $58.00 (même prix)
   • TopPrix: $59.00 (+$1.00)

🎯 DÉCISION : MonMagasin sélectionné !
Pourquoi ?
   • Prix le plus bas : $58.00 (ex-aequo avec 1 autre)
   • Sélectionné en premier par ordre alphabétique
```

**Produit non trouvé** :
```
🔍 Recherche: Achète un casque

❌ Aucun marchand ne propose ce produit en stock

Suggestions :
   • Vérifier l'orthographe
   • Essayer "casque audio" ou "headphones"
```

### Code Clé

**Fichier** : `demo/internal/obs/handler.go`

**Fonction** : `executeBuyingFlow(instruction string)`

**Lignes** : 349-410

### Références

- Commit `3723498` (2026-06-04) : feat: Système multi-agents 3-agents avec messages détaillés
- Fichier : `demo/internal/obs/handler.go` (executeBuyingFlow)
- Lié à : ADR-005 (Agent Acheteur Intégré), ADR-007 (Toast Notifications)
- Interface : http://localhost:9002/arena (panel d'activité)
