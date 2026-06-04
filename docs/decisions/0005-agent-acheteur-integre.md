---
parent: Decisions
nav_order: 5
title: ADR-005 Agent Acheteur Intégré

status: accepted
date: 2026-06-04
decision-makers: Elsa Singer
---

# Agent Acheteur Intégré dans l'Interface Web au lieu d'Agent Gemini Externe

## Contexte et Problème

Pour tester le système de pricing multi-agents, nous avons besoin d'un agent acheteur qui :
- Recherche des produits dans le Shopping Graph
- Compare les prix de tous les marchands
- Sélectionne le marchand le moins cher
- Affiche sa décision de manière transparente

L'implémentation initiale utilisait un agent Gemini externe, mais cela pose des problèmes :
- Setup complexe (GCP, API keys, authentification)
- Script séparé à lancer (`acheter.sh`)
- Pas de feedback visuel en temps réel dans l'interface
- Difficile à démontrer lors d'une présentation

Comment permettre à l'utilisateur de tester l'achat directement depuis l'interface web sans configuration externe ?

## Facteurs de Décision

* **Facilité d'utilisation** : Zero setup, utilisable immédiatement
* **Feedback temps réel** : Voir les étapes de décision en direct
* **Intégration UX** : Expérience fluide dans le dashboard arène
* **Démonstrabilité** : Facile à montrer lors d'une présentation
* **Transparence** : Afficher le raisonnement de l'agent
* **Maintenabilité** : Code simple, pas de dépendances externes

## Options Considérées

* Option 1: Garder agent Gemini externe
* Option 2: Agent côté client (JavaScript)
* Option 3: Agent intégré côté serveur (Go)

## Décision

Option choisie: "**Option 3: Agent intégré côté serveur**", car elle offre zero setup, accès direct au Shopping Graph depuis le serveur, et permet des notifications temps réel via SSE sans complexité d'un vrai LLM.

### Conséquences

* Good, because zero setup (pas besoin GCP, Gemini, API keys)
* Good, because feedback temps réel via SSE (Server-Sent Events)
* Good, because UX fluide (input + bouton directement dans l'arène)
* Good, because facile à démontrer (juste taper et cliquer)
* Good, because notifications visuelles (toast + surbrillance marchand)
* Good, because transparence (messages détaillés de chaque étape)
* Bad, because moins "intelligent" qu'un vrai LLM (pas de NLP)
* Bad, because logique d'achat simple (recherche keyword fixe "casque")
* Bad, because pas d'apprentissage ou personnalisation

### Confirmation

L'implémentation est confirmée par :
- **Interface** : Input + bouton "🛒 Acheter" visible dans http://localhost:9002/arena
- **Fonctionnement** : Taper "Achète un casque" → messages apparaissent en temps réel
- **Code serveur** : `executeBuyingFlow()` dans `demo/internal/obs/handler.go`
- **Tests** : `./scripts/acheter.sh` fonctionne toujours comme backup CLI

### Implémentation

**Architecture** :

```
┌──────────────────────────────────────────────────────┐
│  Interface Web (Arène Dashboard)                     │
│  ┌────────────────────────────────────────────────┐  │
│  │  <input id="command-input">                    │  │
│  │  <button>🛒 Acheter</button>                   │  │
│  └─────────────────┬──────────────────────────────┘  │
│                    │ POST /command                    │
└────────────────────┼──────────────────────────────────┘
                     ▼
┌──────────────────────────────────────────────────────┐
│  Serveur Obs-Hub (Go)                                │
│  ┌────────────────────────────────────────────────┐  │
│  │  executeBuyingFlow()                           │  │
│  │  1. POST /search → Shopping Graph              │  │
│  │  2. Parse résultats, trouve le moins cher     │  │
│  │  3. POST /checkout avec AUTO_COMPETE           │  │
│  │  4. Envoie événements SSE                      │  │
│  └────────────────┬───────────────────────────────┘  │
└───────────────────┼──────────────────────────────────┘
                    │ SSE events
                    ▼
┌──────────────────────────────────────────────────────┐
│  Interface Web (Panel + Toast)                       │
│  • 🔍 Recherche: Achète un casque                    │
│  • 📊 Comparaison des prix                           │
│  • 🎯 DÉCISION : MonMagasin sélectionné !            │
│  • 🛒 Création du panier...                          │
│  • ✅ Achat confirmé !                               │
│  • Toast notification en haut à droite               │
│  • Marchand surligné en vert                         │
└──────────────────────────────────────────────────────┘
```

**Fichiers clés** :
- `demo/internal/obs/handler.go` : Fonction `executeBuyingFlow()`
- `demo/internal/obs/dashboard_arena.go` : Input + bouton + handler SSE

## Avantages et Inconvénients des Options

### Option 1: Agent Gemini Externe

Agent LLM indépendant connecté via API Gemini.

* Good, because vrai LLM avec capacités NLP
* Good, because peut comprendre requêtes complexes ("laptop pas cher livré rapidement")
* Good, because peut apprendre et s'améliorer
* Bad, because setup complexe (GCP account, API key, auth)
* Bad, because coût (requêtes Gemini facturées)
* Bad, because script séparé à lancer (`acheter.sh`)
* Bad, because pas de feedback visuel dans l'interface
* Bad, because difficile à démontrer (setup requis)

### Option 2: Agent Côté Client (JavaScript)

Logique d'achat implémentée en JavaScript dans le navigateur.

* Good, because pas de serveur nécessaire
* Good, because réactivité immédiate
* Good, because code simple (fetch API)
* Bad, because pas d'accès Shopping Graph (CORS)
* Bad, because pas d'accès backend Arena (auth)
* Bad, because logique métier dans le client (sécurité)
* Bad, because difficile de simuler un vrai checkout

### Option 3: Agent Intégré Serveur (Choisi)

Fonction Go côté serveur qui simule un agent acheteur.

* Good, because zero setup (fonctionne immédiatement)
* Good, because accès direct Shopping Graph et Arena (même serveur)
* Good, because SSE pour feedback temps réel
* Good, because peut créer de vrais checkouts avec AUTO_COMPETE
* Good, because facile à démontrer (juste taper et cliquer)
* Good, because notifications visuelles (toast + surbrillance)
* Neutral, because logique simple mais suffisante pour la démo
* Bad, because pas de NLP (keyword fixe "casque")
* Bad, because pas aussi "intelligent" qu'un LLM

## Informations Complémentaires

### Flux Utilisateur

```
1. Utilisateur ouvre http://localhost:9002/arena
2. Tape dans l'input : "Achète un casque"
3. Clique "🛒 Acheter"
   ↓
4. Serveur exécute executeBuyingFlow()
   • Envoie SSE: "🔍 Recherche: Achète un casque"
   • POST /search au Shopping Graph
   • Parse résultats, trouve le moins cher
   • Envoie SSE: "📊 Comparaison des prix : ..."
   • Envoie SSE: "🎯 DÉCISION : MonMagasin sélectionné !"
   • POST /checkout avec AUTO_COMPETE
   • Envoie SSE: "✅ Achat confirmé !"
   • Envoie SSE type merchant_selected
   ↓
5. Interface affiche en temps réel :
   • Messages dans panel d'activité
   • Toast notification "🎯 Marchand sélectionné !"
   • Marchand surligné en vert dans la liste
```

### Code Clé

**Fonction serveur** (`handler.go`) :
```go
func (h *Handler) executeBuyingFlow(instruction string) {
    // 1. Recherche Shopping Graph
    h.hub.Add(Event{
        Type: "agent_message",
        Data: map[string]any{"message": "🔍 Recherche: " + instruction},
    })
    
    searchResp, _ := http.Post(h.graphURL+"/search", ...)
    
    // 2. Trouve le moins cher
    var cheapest *struct{ MerchantID, MerchantName string; Price int }
    for _, r := range results.Results {
        if r.InStock && (cheapest == nil || r.Price < cheapest.Price) {
            cheapest = &struct{...}{r.MerchantID, r.MerchantName, r.Price}
        }
    }
    
    // 3. Envoie comparaison et décision
    h.hub.Add(Event{Type: "agent_message", Data: comparison})
    h.hub.Add(Event{Type: "agent_message", Data: decision})
    
    // 4. Crée checkout
    http.Post(h.arenaURL+"/"+cheapest.MerchantID+"/checkouts", ...)
    
    // 5. Notification finale
    h.hub.Add(Event{Type: "merchant_selected", Data: cheapest})
}
```

**Handler JavaScript** (`dashboard_arena.go`) :
```javascript
if (ev.type === 'agent_message' && ev.data && ev.data.message) {
    appendToPanel('result-entry', '🤖 Agent', ev.data.message);
}

if (ev.type === 'merchant_selected' && ev.data) {
    showToast(
        '🎯 Marchand sélectionné !',
        ev.data.merchant_name + ' remporte la vente',
        'success'
    );
    highlightMerchant(ev.data.merchant_name);
}
```

### Exemple Concret

**Input utilisateur** : "Achète un casque"

**Messages affichés** :
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

🏆 Gagnant : MonMagasin sélectionné !
```

**Toast notification** : "🎯 Marchand sélectionné ! MonMagasin remporte la vente avec le meilleur prix : $51.30"

**Effet visuel** : MonMagasin surligné en vert dans la liste des marchands

### Backup CLI

Le script `./scripts/acheter.sh` reste disponible comme alternative CLI pour :
- Tests automatisés
- Démo en ligne de commande
- Situations où l'interface web n'est pas accessible

### Références

- Commit `3723498` (2026-06-04) : feat: Système multi-agents 3-agents avec interface arène interactive
- Fichiers : `demo/internal/obs/handler.go`, `demo/internal/obs/dashboard_arena.go`
- Lié à : ADR-007 (Toast Notifications), ADR-006 (Messages Détaillés)
