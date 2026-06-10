---
parent: Decisions
nav_order: 2
title: ADR-002 Multi-Transport Architecture

status: accepted
date: 2026-03-11
decision-makers: Olivier Wulveryck
---

# Architecture Multi-Transport pour Supporter Web, LLM et Agents

## Context and Problem Statement

Le projet UCP merchant test doit être accessible par **trois types de clients** avec des besoins fondamentalement différents :

1. **Clients Web/Mobile** : Applications classiques (REST)
2. **LLM Clients** : Claude Desktop, IDEs (MCP - Model Context Protocol)  
3. **Agents Autonomes** : Shopping Graph, Client Agent (A2A avec OAuth2)

Comment exposer la même logique métier à ces trois types de clients sans duplication de code tout en respectant leurs protocoles spécifiques ?

## Decision Drivers

* **Zero duplication** : La logique métier ne doit être implémentée qu'une fois
* **Conformité specs** : Respecter UCP (REST), MCP (tools), et A2A (OAuth2 + discovery)
* **Intégration Web** : Support des clients HTTP classiques (web, mobile, tests)
* **Intégration LLM** : Exposer les capabilities comme tools pour Claude Desktop
* **Communication Agent-to-Agent** : Discovery automatique, auth sécurisée, session management
* **Extensibilité** : Pouvoir ajouter de nouveaux transports sans refonte
* **Maintenabilité** : Chaque transport doit être un adaptateur pur

## Considered Options

* Option 1: REST Uniquement
* Option 2: MCP Uniquement  
* Option 3: A2A Uniquement
* Option 4: Architecture Multi-Transport (REST + MCP + A2A)

## Decision Outcome

Chosen option: "**Option 4: Architecture Multi-Transport**", because it's the only option that meets all decision drivers. Each client type uses the transport protocol best suited to its needs, while all three delegate to the same `merchant.Merchant` interface with zero business logic duplication.

### Consequences

* Good, because chaque client utilise le protocole adapté à ses besoins
* Good, because zero duplication de logique métier (tous délèguent à `merchant.Merchant`)
* Good, because extensible (pattern établi pour ajouter GraphQL, gRPC, etc.)
* Good, because conformité totale aux specs UCP, MCP et A2A
* Good, because ajout A2A sans modifier business logic valide l'architecture
* Bad, because complexité accrue (3 packages ~3300 lignes)
* Bad, because maintenance de 3 transports (60 tests UCP + 43 MCP + tests A2A)

### Confirmation

L'architecture est confirmée par :
- **Tests de conformance** : 60 tests UCP passent (REST), 43 tests MCP passent
- **Zero duplication** : `git grep "CreateCheckout"` montre que tous les transports appellent `h.merchant.CreateCheckout()`
- **Extensibilité validée** : Ajout A2A (commit `13e9206`) sans modifier core business logic

### Implementation

Architecture en Couches

```
┌─────────────────────────────────────────────────────────┐
│              3 Transport Adapters                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   REST       │  │     MCP      │  │     A2A      │  │
│  │   HTTP       │  │  JSON-RPC    │  │  JSON-RPC    │  │
│  │              │  │  Tools       │  │  OAuth2      │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │           │
└─────────┼─────────────────┼─────────────────┼───────────┘
          └─────────────────┴─────────────────┘
                            │
          ┌─────────────────▼──────────────────┐
          │   merchant.Merchant interface      │
          │   • Cataloger  (catalog)           │
          │   • Carter     (cart)              │
          │   • Checkouter (checkout)          │
          │   • Orderer    (order)             │
          └────────────────────────────────────┘
                            │
          ┌─────────────────▼──────────────────┐
          │   sample_implementation/           │
          │   (Business Logic)                 │
          └────────────────────────────────────┘
```

### Les 3 Transports

## Transport 1 : REST

**Use Case** : Clients web, mobile, tests de conformance, debug

**Caractéristiques** :
- Endpoints HTTP classiques (`POST /checkout`, `PATCH /checkout/:id`, etc.)
- Stateless (pas de session)
- Content-Type: `application/json`
- Auth : Simulation endpoint (API key)

**Implémentation** : `pkg/merchant/transport/rest/` (~400 lignes)

**Endpoints principaux** :
```
POST   /checkout                 # Create checkout
PATCH  /checkout/:id             # Update checkout
GET    /checkout/:id             # Get checkout
POST   /checkout/:id/order       # Create order
DELETE /checkout/:id             # Cancel checkout
```

**Utilisé par** :
- Tests de conformance UCP (60 tests Python)
- Debug avec curl
- Clients web/mobile classiques

---

## Transport 2 : MCP (Model Context Protocol)

**Use Case** : Intégration avec LLM clients (Claude Desktop, IDEs)

**Caractéristiques** :
- JSON-RPC 2.0
- Tool definitions typées (schema JSON)
- StreamableHTTP transport
- Library : `mark3labs/mcp-go` v0.45.0

**Implémentation** : `pkg/merchant/transport/mcp/` (~900 lignes)

**Tools exposés** :
```json
{
  "tools": [
    {
      "name": "catalog_list_products",
      "description": "List all products in the catalog",
      "inputSchema": { ... }
    },
    {
      "name": "checkout_create",
      "description": "Create a new checkout session",
      "inputSchema": { ... }
    }
  ]
}
```

**Format requête** :
```json
POST /mcp
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "id": "1",
  "params": {
    "name": "checkout_create",
    "arguments": {
      "items": [{"product_id": "prod_1", "quantity": 2}]
    }
  }
}
```

**Utilisé par** :
- Claude Desktop (MCP servers)
- IDEs avec support MCP
- LLM function calling

---

## Transport 3 : A2A (Agent-to-Agent Protocol)

**Use Case** : Communication agent-to-agent autonome (Shopping Graph, Client Agent)

**Caractéristiques** :
- JSON-RPC 2.0
- OAuth2 + PKCE (authentification robuste)
- Discovery via `/.well-known/agent` (agent card)
- Session management (session_id, conversation_id)

**Implémentation** :
- Client : `demo/internal/a2aclient/` (~400 lignes)
- Server : `pkg/merchant/transport/a2a/` (~2000 lignes)

**Flow complet** :

```
1. Discovery (no auth required)
   GET /.well-known/agent
   → agent_card.json (capabilities, auth endpoints)

2. OAuth2 + PKCE Flow
   GET /oauth2/authorize?code_challenge=...
   → authorization code
   
   POST /oauth2/token
   { code, code_verifier }
   → { access_token, expires_in: 3600 }

3. A2A Actions (authenticated)
   POST /a2a
   Authorization: Bearer <token>
   {
     "jsonrpc": "2.0",
     "method": "message/send",
     "params": {
       "session_id": "sess-abc",
       "message": {
         "parts": [{
           "kind": "data",
           "data": {
             "action": "checkout.create",
             "items": [...]
           }
         }]
       }
     }
   }
```

**Actions supportées** (16 handlers) :
- `catalog.*`, `cart.*`, `checkout.*`, `order.*`

**Utilisé par** :
- Shopping Graph (polling catalogues)
- Client Agent (checkout/order)

---

## Pros and Cons of the Options

### Option 1: REST Uniquement

Exposer uniquement des endpoints HTTP REST classiques.

* Good, because simple et bien connu
* Good, because tooling mature (curl, Postman, etc.)
* Good, because conforme UCP
* Bad, because LLM clients doivent parser REST (pas de tool definitions)
* Bad, because pas de discovery automatique pour agents
* Bad, because auth custom à implémenter pour agents

---

### Option 2: MCP Uniquement

Exposer uniquement des tools MCP via JSON-RPC 2.0.

* Good, because tool definitions typées pour LLM (schema JSON)
* Good, because JSON-RPC 2.0 standard
* Good, because excellent intégration Claude Desktop
* Bad, because pas d'auth (local seulement, pas de OAuth2)
* Bad, because sémantique LLM→Tool, pas adaptée pour Agent→Service
* Bad, because clients web doivent implémenter JSON-RPC
* Neutral, because conforme MCP mais pas UCP

---

### Option 3: A2A Uniquement

Exposer uniquement A2A avec OAuth2 + PKCE + discovery.

* Good, because discovery automatique via agent card
* Good, because OAuth2 + PKCE (auth robuste)
* Good, because session management natif
* Good, because excellent pour agents autonomes
* Bad, because over-engineering pour clients web simples
* Bad, because LLM clients doivent implémenter OAuth2
* Bad, because complexité debug (curl nécessite OAuth flow complet)

---

### Option 4: Architecture Multi-Transport (Chosen)

Implémenter les 3 transports en parallèle, tous déléguant à `merchant.Merchant`.

* Good, because chaque client utilise le protocole qui lui convient
* Good, because zero duplication (logique métier unique dans `merchant.Merchant`)
* Good, because extensible (pattern clair pour ajouter GraphQL, gRPC)
* Good, because conformité totale (UCP, MCP, A2A)
* Good, because validation architecture (ajout A2A sans modifier core)
* Bad, because 3 implémentations de sérialisation (~3300 lignes)
* Bad, because complexité maintenance (3 transports à maintenir)
* Bad, because tests multipliés (60 UCP + 43 MCP + tests A2A)

---

## More Information

### Tableau Comparatif des Transports

| Critère | REST | MCP | A2A |
|---------|------|-----|-----|
| **Use Case** | Web/Mobile | LLM | Agents |
| **Protocol** | HTTP REST | JSON-RPC 2.0 | JSON-RPC + OAuth2 |
| **Auth** | Custom | ❌ | ✅ OAuth2+PKCE |
| **Discovery** | ❌ | Tool list | ✅ Agent card |
| **Session** | ❌ | Via conv | ✅ session_id |
| **LOC** | ~400 | ~900 | ~2000 |
| **Tests** | 60 UCP | 43 JSON-RPC | Unit tests |
| **Utilisé** | Web, curl | Claude Desktop | Graph, Agent |

### Impact Architecture

**Commit `eb9db23` (10 mars)** : Extraction transports
- +2742/-2600 lignes
- Création `pkg/merchant/transport/{rest,mcp}/`
- Interface `merchant.Merchant` créée

**Commit `13e9206` (11 mars)** : Ajout A2A
- +2002 lignes (20 fichiers)
- Zéro modification core business logic
- Validation du pattern

### Code Zero-Duplication

```go
// Tous les transports appellent la même interface
checkout, err := h.merchant.CreateCheckout(ctx, req)
```

**Résultat** : 0 duplication de logique métier.

### Références

### Spécifications
- Universal Commerce Protocol (UCP) : https://universalcommerceprotocol.org
- Model Context Protocol (MCP) : https://modelcontextprotocol.io
- A2A Protocol : https://a2a.dev

### ADR Liés
- [ADR 001](001-multi-agent-shopping-architecture.md) : Architecture Multi-Agent

### Commits Clés
- `eb9db23` (2026-03-10) : Extract transport packages
- `eeaf043` (2026-03-10) : Replace custom MCP with mcp-go
- `13e9206` (2026-03-11) : Add A2A transport
