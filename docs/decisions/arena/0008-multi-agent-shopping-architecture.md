# ADR 001 : Architecture Multi-Agent pour Shopping Demo

- **Date** : 2026-03-11
- **Statut** : Accepté
- **Décideurs** : Olivier Wulveryck

## Contexte

Le projet UCP merchant test implémente avec succès le protocole Universal Commerce Protocol (UCP) avec support REST, MCP, et A2A transports. Cependant, pour démontrer l'intérêt des protocoles A2A (Agent-to-Agent) et MCP (Model Context Protocol) dans un contexte réaliste, il manque :

1. **Un cas d'usage concret** : Les tests de conformance valident la spec, mais ne montrent pas comment un agent autonome utiliserait l'API.
2. **Une architecture multi-merchant** : Dans le commerce réel, un acheteur compare plusieurs marchands avant d'acheter.
3. **Un agent orchestrateur** : Un LLM qui coordonne la recherche cross-merchant, la comparaison de prix, et l'achat final.
4. **De l'observabilité** : Visualiser le raisonnement de l'agent et les interactions A2A/MCP.

### Objectifs de la démo

- Démontrer la **valeur des protocoles A2A et MCP** pour l'agentic commerce
- Montrer comment un **agent LLM autonome** peut orchestrer un parcours d'achat complet
- Valider que l'architecture multi-transport (REST, MCP, A2A) supporte réellement des cas d'usage multi-agents

## Décision

Créer un module `demo/` séparé implémentant une **architecture multi-agent** avec 4 composants principaux :

```
┌─────────────────────────────────────────────────────────┐
│                    Client Agent                          │
│              (Gemini/Vertex AI + 8 tools)                │
└───────────┬────────────────────────────┬─────────────────┘
            │                            │
            │ A2A/MCP                    │ SSE events
            │                            │
    ┌───────▼───────┐            ┌──────▼──────┐
    │ Shopping Graph│            │ Obs Hub     │
    │  (:9000)      │            │  (:9002)    │
    └───┬───┬───┬───┘            └─────────────┘
        │   │   │
    A2A │   │   │ A2A
        │   │   │
  ┌─────▼───▼───▼─────┐
  │ 3 Merchant Servers│
  │ (A2A + REST + MCP)│
  │ :8182, :8183, :8184
  └───────────────────┘
```

### Composant 1 : Shopping Graph (Service de Recherche Cross-Merchant)

**Responsabilité** : Agréger et indexer les produits de tous les marchands pour offrir une recherche unifiée.

**Fonctionnement** :
- Polling A2A : Récupère périodiquement les catalogues via A2A protocol (`catalog.list_products`)
- Indexation : Construit un index de recherche inversé (keyword → [products])
- Matching : Utilise **Jaccard similarity** pour matcher la requête utilisateur aux produits
- Ranking : Algorithme configurable pour classer les résultats

**Technologies** :
- A2A client library custom (OAuth2 + PKCE, JSON-RPC message/send)
- Configuration YAML (liste des merchants, URL A2A discovery)

**Port** : 9000

### Composant 2 : Client Agent (LLM Orchestrateur)

**Responsabilité** : Agent autonome Gemini qui exécute le parcours d'achat de bout en bout.

**Capacités** (8 tools via function calling) :
1. `search_products` : Recherche via Shopping Graph
2. `create_checkout` : Crée un checkout chez un merchant (A2A)
3. `update_checkout` : Applique discount codes, sélectionne shipping (A2A)
4. `get_checkout` : Récupère l'état du checkout (A2A)
5. `create_order` : Finalise la commande (A2A)
6. `cancel_checkout` : Annule un checkout (A2A)
7. `list_merchants` : Liste des marchands disponibles
8. `emit_thinking` : Envoie les étapes de raisonnement à l'Obs Hub

**Algorithme** :
1. **Search** : Interroge Shopping Graph pour trouver les produits correspondants
2. **Create checkouts** : Crée des checkouts chez 2-3 marchands potentiels (parallèle via A2A)
3. **Apply discounts** : Teste les codes promo suggérés (hints dans les métadonnées produits)
4. **Compare prices** : Calcule le prix total final (item + shipping + discounts)
5. **Place order** : Achète chez le marchand le moins cher
6. **Cleanup** : Annule les checkouts non utilisés

**Technologies** :
- Gemini (via Vertex AI) avec function calling
- A2A client pour communication merchant
- SSE client pour envoyer les events à Obs Hub

### Composant 3 : Observability Hub (Dashboard Temps Réel)

**Responsabilité** : Dashboard web pour visualiser le comportement de l'agent en temps réel.

**Fonctionnalités** :
- Event stream SSE (Server-Sent Events) : `agent_thinking`, `tool_call`, `tool_result`
- Interface web HTML/JS vanilla (pas de framework)
- Affichage chronologique des actions de l'agent

**Port** : 9002

### Composant 4 : Trois Instances Merchant

**Configuration** :
- **SuperShop** (:8182) : 6 produits, codes `SAVE10` (10%), `WELCOME15` (15%), free shipping >= $100
- **MegaMart** (:8183) : 5 produits, code `MEGA10` (10%), free shipping >= $150
- **BudgetBuy** (:8184) : 5 produits, codes `BUDGET20` (20%), `SAVE5` ($5 off), free shipping >= $80

Chaque merchant expose :
- A2A endpoint (pour le Shopping Graph et Client Agent)
- REST API (pour debug)
- MCP server (pour intégration LLM alternative)

## Choix Techniques

### Jaccard Similarity pour le Matching

Le Shopping Graph utilise Jaccard pour matcher requêtes → produits :

```
Jaccard(A, B) = |A ∩ B| / |A ∪ B|
```

- **A** : Ensemble de mots-clés de la requête (après tokenization)
- **B** : Ensemble de mots-clés du produit (name + description + category)

Seuil de pertinence : 0.2 (empirique). Simple, rapide, sans dépendances ML.

### Go Workspace pour Multi-Module

Le projet utilise un **workspace Go** (`go.work`) pour gérer 2 modules :
- Module racine : `github.com/owulveryck/ucp-merchant-test`
- Module demo : `github.com/owulveryck/ucp-merchant-test/demo`

Cela permet :
- Demo d'importer `pkg/merchant`, `pkg/catalog`, `pkg/model` depuis le module racine
- Tests cross-module dans le même repo
- Build indépendant de chaque module

### A2A Client Library Custom

Pas de SDK A2A officiel disponible → implémentation custom dans `demo/internal/a2aclient/` :
- **OAuth2 + PKCE** : Flow complet (discovery → authorize → token exchange)
- **JSON-RPC transport** : Fonctions `message()` et `send()` selon spec A2A
- **Session management** : Gestion du `session_id` et `conversation_id`

### Gemini via Vertex AI (pas Anthropic)

Choix de **Gemini** plutôt que Claude pour le Client Agent :
- Projet initialement déployé sur Google Cloud (cohérence avec l'infra)
- Function calling robuste et bien documenté
- Coût inférieur pour les appels répétés (polling, comparaison prix)

**Note** : L'architecture supporte n'importe quel LLM avec function calling. Un agent Claude pourrait remplacer Gemini sans modification structurelle.

## Conséquences

### Positives

- ✅ **Démo convaincante** : Parcours d'achat autonome end-to-end (search → checkout → order)
- ✅ **Validation multi-transport** : L'architecture avec adaptateurs protocolaires supporte réellement le multi-agent
- ✅ **Réutilisabilité** : Move `internal/` → `pkg/` permet au demo d'importer les packages
- ✅ **Observabilité** : Dashboard temps réel pour comprendre le raisonnement de l'agent
- ✅ **Extensibilité** : Facile d'ajouter de nouveaux merchants ou agents

### Négatives

- ❌ **Complexité** : 4 binaires à lancer (merchants x3 + shopping-graph + obs-hub + client)
- ❌ **Dépendances GCP** : Vertex AI requis (pas de fallback local)
- ❌ **Coût** : Chaque run consomme des tokens Gemini (mitigé par le choix de Gemini vs Claude)
- ❌ **Latence** : Appels A2A séquentiels + LLM function calling → 10-30s pour un achat

### Risques

- **Gemini instability** : Si Vertex AI est down, la démo ne fonctionne pas. Mitigation : support MCP permettrait un fallback vers Claude Desktop.
- **A2A spec changes** : Si le protocole A2A évolue, la custom library doit être mise à jour.

## Fichiers Concernés

### Créés

| Fichier/Dossier | Description |
|-----------------|-------------|
| `demo/` | Module Go séparé pour la démo |
| `demo/cmd/client/` | Client Agent Gemini (binary) |
| `demo/cmd/shopping-graph/` | Shopping Graph server (binary) |
| `demo/cmd/obs-hub/` | Observability Hub (binary) |
| `demo/internal/a2aclient/` | A2A client library (auth + transport) |
| `demo/internal/client/` | Agent logic + tools |
| `demo/internal/shoppinggraph/` | Graph, poller, matcher, search |
| `demo/internal/obs/` | SSE hub + dashboard HTML |
| `demo/data/merchant_{a,b,c}/` | 3 jeux de données merchants |
| `demo/config/shopping_graph.yaml` | Config merchants pour le graph |
| `demo/scripts/run_demo.sh` | Launcher pour démarrer tous les services |
| `go.work` | Workspace Go multi-module |

### Modifiés

| Fichier | Modification |
|---------|-------------|
| `internal/*` → `pkg/*` | Move pour permettre import par demo/ (commit `6dc541d`) |
| `sample_implementation/main.go` | Ajout flag `--merchant-name` pour différencier les instances |

## Évolutions Post-Implémentation

### Arena Mode (2026-03-13, commit c25ee9f)

**Contexte** : La démo initiale lance 3 merchants statiques. Pour une démo conférence, besoin d'un mode **compétition multi-tenant** où plusieurs merchants peuvent ajuster dynamiquement prix, stock, et discounts.

**Ajouts** :
- **Arena server** (`demo/cmd/arena/`) : Multi-tenant merchant spawner avec API de configuration
- **Merchant dashboards** : Tableau de bord par merchant avec activity log (browsing, checkouts, sales)
- **Landing page** : Interface de sélection merchant pour la démo
- **Enhanced observability** : Events `tool_call`, `tool_result`, `agent_thinking` dans l'arena monitor

**Impact** : +3971 lignes. L'arena devient le mode principal pour les démos publiques.

### Arena Ranking Algorithm (2026-04-03, commit 5d82405)

**Problème** : Le ranking initial du Shopping Graph était dominé par le **bid** (enchères publicitaires), créant une mauvaise UX où les produits chers mais sponsorisés apparaissaient en premier.

**Solution** : Nouvel algorithme **RankArena** avec pondération équilibrée :

```
Score = Prix (5 pts) + Stock (2 pts) + Bid (3 pts)
```

- **Prix** : Inversement proportionnel (prix bas = score élevé)
- **Stock** : +2 si disponible, 0 si rupture
- **Bid** : Influence réduite (3 pts max au lieu de dominer)

**Impact** : Le client agent achète maintenant réellement les produits les moins chers, pas les plus sponsorisés.

### Buying Modes (2026-05-21, commit 1a482f9)

**Contexte** : Les utilisateurs ont parfois des contraintes de **délai** (livraison rapide) plutôt que de prix.

**Ajout de 2 modes détectés automatiquement** :
1. **Mode "Moins cher"** (par défaut) : Sélectionne l'option de livraison la moins chère, achète au prix total le plus bas
2. **Mode "Plus rapide"** : Détecte les mots-clés (`rapidement`, `express`, `vite`) → sélectionne l'option la plus rapide, achète au délai le plus court

**Détection des délais** :
- "Express", "Expedited" → 1-2 jours
- "Standard", "Regular" → 3-5 jours

**Impact** : L'agent s'adapte au contexte utilisateur. Exemple :
- "Achète des fleurs" → mode moins cher
- "Achète des fleurs rapidement" → mode plus rapide (accepte surcoût)

### Retour d'expérience (audit mai 2026)

**Ce qui fonctionne bien :**
- L'architecture multi-agent est robuste et extensible (ajout d'agents/merchants sans refonte)
- Le Shopping Graph avec Jaccard similarity est suffisamment précis pour la démo (pas besoin de ML complexe)
- L'observabilité temps réel est un atout majeur pour les présentations
- La séparation agent/graph/hub permet des évolutions indépendantes

**Axes d'amélioration identifiés :**
- **Latence** : Le polling A2A du Shopping Graph introduit de la latence (mitigé par un cache, non implémenté)
- **Tests automatisés** : Aucun test end-to-end pour le parcours agent complet (difficile à mocker)
- **Configuration** : Les 3 datasets merchants sont hardcodés (pas de générateur dynamique)
- **Parallélisme agent** : Le Client Agent fait des appels A2A séquentiels alors que certains pourraient être parallèles (checkouts multiples)

## Références

### Protocoles et Spécifications
- Universal Commerce Protocol (UCP) : https://universalcommerceprotocol.org
- A2A Protocol : https://a2a.dev
- Model Context Protocol (MCP) : https://modelcontextprotocol.io

### ADR Liés
- [ADR 002](002-a2a-protocol-inter-agent-communication.md) : Protocole A2A pour Communication Inter-Agents

### Commits Clés
- `6dc541d` (2026-03-11) : Add multi-agent shopping demo, move internal/ to pkg/
- `c25ee9f` (2026-03-13) : Add Arena mode with merchant dashboards and observability
- `5d82405` (2026-04-03) : Add arena2 dashboard and arena ranking algorithm
- `1a482f9` (2026-05-21) : Update dashboard UI to reflect 2 buying modes
