# ADR 002 : Protocole A2A pour Communication Inter-Agents

- **Date** : 2026-03-11
- **Statut** : Accepté
- **Décideurs** : Olivier Wulveryck
- **Lié à** : ADR 001 (Architecture Multi-Agent)

## Contexte

L'architecture multi-agent (ADR 001) nécessite des communications machine-to-machine entre :
1. **Shopping Graph → Merchants** : Polling périodique des catalogues (A2A `catalog.list_products`)
2. **Client Agent → Merchants** : Interactions checkout/order (A2A `checkout.create`, `order.create`, etc.)
3. **Client Agent → Shopping Graph** : Recherche de produits cross-merchant

### Exigences fonctionnelles

- **Découverte dynamique** : Le Shopping Graph doit pouvoir découvrir les endpoints et capabilities des merchants
- **Authentification** : Les agents doivent s'authentifier auprès des merchants
- **Actions structurées** : Support d'opérations complexes (checkout avec items, addresses, discounts)
- **Gestion de session** : Certaines interactions nécessitent un contexte conversationnel

### Exigences non-fonctionnelles

- ✅ **Interopérabilité** : N'importe quel agent compatible doit pouvoir communiquer avec n'importe quel merchant
- ✅ **Standardisation** : Protocole ouvert, documenté, avec une spec claire
- ✅ **Sécurité** : Authentification robuste (pas de simple API keys)
- ✅ **Extensibilité** : Support de nouvelles actions sans breaking changes
- ✅ **Simplicité** : Implémentation client raisonnablement simple (pas de stack complexe)

## Décision

Adopter le **Agent-to-Agent Protocol (A2A)** comme protocole de communication inter-agents.

### Qu'est-ce que A2A ?

A2A est un protocole standardisé pour la communication agent-to-agent défini par [a2a.dev](https://a2a.dev). Il spécifie :

1. **Discovery** : Endpoint `/.well-known/agent` retournant les capabilities et auth methods
2. **Authentication** : OAuth2 avec PKCE (Proof Key for Code Exchange)
3. **Transport** : JSON-RPC 2.0 avec méthodes `message/send` et `message/list`
4. **Actions** : Format `action` + `data` dans les message parts
5. **Session management** : `session_id` et `conversation_id` pour le contexte

### Architecture de communication

```
┌──────────────────┐
│  Shopping Graph  │
│   (A2A Client)   │
└────────┬─────────┘
         │ 1. Discovery (GET /.well-known/agent)
         │ 2. OAuth2 flow (authorize → token)
         │ 3. Polling (POST /a2a/message/send)
         ▼
┌────────────────────────────────┐
│   Merchant A2A Endpoint        │
│  pkg/merchant/transport/a2a/   │
└────────────────────────────────┘
         │
         ▼ délègue à
┌────────────────────────────────┐
│   merchant.Merchant interface  │
└────────────────────────────────┘
```

### Implémentation

**Client A2A** (`demo/internal/a2aclient/`) :
- **Auth** : OAuth2 + PKCE flow complet (discovery → authorize → token exchange)
- **Token caching** : Cache les tokens avec expiration automatique
- **JSON-RPC** : Méthode `message/send` avec actions structurées
- ~400 lignes (auth.go + client.go + types.go)

**Server A2A** (`pkg/merchant/transport/a2a/`) :
- **Executor** : Dispatch des actions vers les handlers
- **16 action handlers** : catalog.*, cart.*, checkout.*, order.*
- **Session tracking** : Gestion des `session_id` par merchant
- **Discovery** : Agent card JSON avec capabilities et auth endpoints
- ~2000 lignes (+20 fichiers, tests inclus)

## Alternatives Considérées

### Alternative 1 : REST API Direct

**Pour** :
- ✅ Simple, bien connu
- ✅ Tooling mature (curl, Postman, OpenAPI)
- ✅ Déjà implémenté (`pkg/merchant/transport/rest/`)

**Contre** :
- ❌ Pas de standard pour discovery (chaque API documente différemment)
- ❌ Pas de session management built-in
- ❌ Authentification hétérogène (API keys, Basic Auth, OAuth2 custom)
- ❌ Pas conçu pour agent-to-agent (conçu pour human-to-service)

**Verdict** : Rejeté. REST est excellent pour les clients web/mobile, mais manque de standardisation pour les communications agent-to-agent.

---

### Alternative 2 : gRPC

**Pour** :
- ✅ Performant (binaire, HTTP/2)
- ✅ Typed contracts (protobuf)
- ✅ Streaming bidirectionnel

**Contre** :
- ❌ Nécessite proto definitions partagées (couplage)
- ❌ Pas de discovery standard (besoin de service registry)
- ❌ Complexité tooling (protoc, generation)
- ❌ Pas de spec pour auth agent-to-agent

**Verdict** : Rejeté. Over-engineering pour notre use case. Le typage protobuf est un avantage pour des systèmes à haute performance, mais ajoute de la complexité sans bénéfice clair ici.

---

### Alternative 3 : Message Queue (RabbitMQ, Kafka)

**Pour** :
- ✅ Asynchrone (découplage temporel)
- ✅ Scalable (distribution de charge)
- ✅ Resilience (retry, dead-letter)

**Contre** :
- ❌ Complexité opérationnelle (broker à déployer/maintenir)
- ❌ Over-engineering pour polling synchrone
- ❌ Pas de standard pour discovery/auth
- ❌ Latence accrue pour les opérations synchrones (checkout/order)

**Verdict** : Rejeté. Les message queues sont excellentes pour les architectures event-driven à grande échelle, mais notre démo n'a pas ces besoins. Le Shopping Graph fait du polling simple, et le Client Agent a besoin de réponses synchrones.

---

### Alternative 4 : Model Context Protocol (MCP)

**Pour** :
- ✅ Déjà implémenté (`pkg/merchant/transport/mcp/`)
- ✅ Conçu pour LLM-tool interaction
- ✅ JSON-RPC 2.0 transport

**Contre** :
- ❌ **Sémantique différente** : MCP est LLM → Tool, pas Agent → Service
- ❌ Pas de session management multi-turn
- ❌ Discovery limité (liste de tools, pas de capabilities)
- ❌ Pas conçu pour polling (Shopping Graph cas d'usage)

**Verdict** : Rejeté. MCP est excellent pour exposer des tools à Claude Desktop ou d'autres LLM clients, mais A2A est mieux adapté pour les communications agent-to-agent autonomes.

**Note** : Le projet supporte **les deux** (MCP et A2A) pour différents use cases :
- **MCP** : Pour intégration avec LLM clients (Claude Desktop, IDEs)
- **A2A** : Pour communications inter-agents (Shopping Graph, Client Agent)

---

## Trade-offs

### Positifs

✅ **Standardisation** : Spec ouverte (a2a.dev), interopérabilité garantie  
✅ **Sécurité** : OAuth2 + PKCE (pas de credentials hardcodés)  
✅ **Discovery** : Agent card JSON (capabilities, auth endpoints)  
✅ **Session management** : Built-in via `session_id`, `conversation_id`  
✅ **Extensibilité** : Nouvelles actions sans breaking protocol  
✅ **Multi-transport** : S'intègre proprement dans l'architecture existante (l'architecture multi-transport du projet)  

### Négatifs

❌ **Custom implementation** : Pas de SDK Go officiel → implémentation ~400 lignes client + ~2000 lignes server  
❌ **Complexité OAuth2** : Flow PKCE plus complexe qu'API keys (mais plus sécurisé)  
❌ **Latence** : Overhead JSON-RPC + OAuth token refresh (mitigé par token caching)  
❌ **Spec en évolution** : A2A est récent (risque de breaking changes)  

### Risques et Mitigations

**Risque 1 : A2A spec changes**  
- Mitigation : Versionning dans l'agent card (`"protocol_version": "1.0"`)
- Mitigation : Tests de conformité A2A (comme UCP conformance tests)

**Risque 2 : Pas de SDK officiel**  
- Mitigation : Notre implémentation est bien testée (~500 lignes de tests)
- Mitigation : Open-source notre client A2A si d'autres projets en ont besoin

**Risque 3 : Overhead performance**  
- Mitigation : Token caching (évite OAuth flow à chaque requête)
- Mitigation : HTTP keep-alive (connexions réutilisées)
- Impact mesuré : ~50-100ms par requête (acceptable pour une démo)

---

## Conséquences

### Impact Architecture

**Commit `13e9206` (11 mars 2026)** : Ajout A2A transport
- +2002 lignes (20 fichiers) : executor, 16 handlers, tests
- S'intègre via l'interface `merchant.Merchant` (validation de l'architecture multi-transport du projet)
- Aucune modification du core business logic

**Commit `6dc541d` (11 mars 2026)** : Client A2A dans demo
- +400 lignes (3 fichiers) : auth.go, client.go, types.go
- Utilisé par Shopping Graph (polling) et Client Agent (checkout/order)

### Validation de l'Architecture Multi-Transport

L'ajout d'A2A **valide le pattern multi-transport** (l'architecture multi-transport du projet) :
- Aucune duplication de logique métier
- Transport ajouté en 1 journée (~2000 lignes)
- Tests unitaires complets (128 + 137 + 242 + 111 + 192 lignes de tests)

### Interopérabilité Démontrée

L'architecture supporte maintenant **3 protocoles** pour différents clients :
- **REST** : Clients web, mobile, debug (curl)
- **MCP** : LLM clients (Claude Desktop, IDEs)
- **A2A** : Agents autonomes (Shopping Graph, Client Agent)

Un même merchant expose les 3 protocoles simultanément, chacun délégant à la même interface `merchant.Merchant`.

---

## Exemples de Communication A2A

### Discovery

```bash
GET https://merchant-a.com/.well-known/agent

Response:
{
  "agent_card": {
    "name": "SuperShop Merchant",
    "capabilities": ["catalog", "cart", "checkout", "order"],
    "auth": {
      "methods": ["oauth2"],
      "oauth2": {
        "authorization_endpoint": "/oauth2/authorize",
        "token_endpoint": "/oauth2/token"
      }
    },
    "endpoints": {
      "a2a": "/a2a"
    }
  }
}
```

### Catalog Request (Shopping Graph → Merchant)

```bash
POST https://merchant-a.com/a2a
Authorization: Bearer <token>

{
  "jsonrpc": "2.0",
  "method": "message/send",
  "id": "req-123",
  "params": {
    "message": {
      "message_id": "msg-123",
      "role": "user",
      "parts": [
        {
          "kind": "data",
          "data": {
            "action": "catalog.list_products",
            "category": "flowers"
          }
        }
      ]
    }
  }
}
```

### Checkout Request (Client Agent → Merchant)

```bash
POST https://merchant-a.com/a2a
Authorization: Bearer <token>

{
  "jsonrpc": "2.0",
  "method": "message/send",
  "id": "req-456",
  "params": {
    "session_id": "session-abc",
    "message": {
      "message_id": "msg-456",
      "role": "user",
      "parts": [
        {
          "kind": "data",
          "data": {
            "action": "checkout.create",
            "items": [{"product_id": "prod_123", "quantity": 2}],
            "buyer": {"customer_id": "cust_1"}
          }
        }
      ]
    }
  }
}
```

---

## Évolutions Futures Possibles

### Tests de Conformité A2A

Comme pour UCP (60 tests de conformance), créer une suite de tests A2A pour valider la conformité du transport :
- Discovery (agent card)
- OAuth2 flow
- Actions catalog/cart/checkout/order
- Session management
- Error handling

### SDK A2A Go Open-Source

Si d'autres projets Go ont besoin d'un client A2A, extraire `demo/internal/a2aclient/` dans un package réutilisable :
- `github.com/owulveryck/a2a-go`
- Documentation complète
- Tests de compatibilité avec la spec A2A

### Support A2A dans d'autres Implémentations UCP

Le transport A2A est réutilisable pour n'importe quelle implémentation de `merchant.Merchant`. Si un autre merchant (e.g., Shopify, WooCommerce) implémente l'interface, il obtient A2A gratuitement.

---

## Références

### Protocoles et Spécifications
- A2A Protocol : https://a2a.dev

### ADR Liés
- [ADR 001](001-multi-agent-shopping-architecture.md) : Architecture Multi-Agent Shopping

### Commits Clés
- `13e9206` (2026-03-11) : Add A2A transport binding to UCP merchant server
- `6dc541d` (2026-03-11) : Add multi-agent shopping demo
- `eb9db23` (2026-03-10) : Extract transport packages (pattern multi-transport)
