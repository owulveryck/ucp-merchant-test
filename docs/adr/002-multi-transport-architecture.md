# ADR 002 : Architecture Multi-Transport (REST, MCP, A2A)

- **Date** : 2026-03-10 - 2026-03-11
- **Statut** : Accepté
- **Décideurs** : Olivier Wulveryck
- **Lié à** : ADR 001 (Architecture Multi-Agent)

## Contexte

Le projet UCP merchant test doit être accessible par **différents types de clients** avec des besoins distincts :

1. **Clients Web/Mobile** : Applications classiques utilisant des API REST
2. **LLM Clients** : Claude Desktop, IDEs avec support MCP (Model Context Protocol)
3. **Agents Autonomes** : Shopping Graph, Client Agent (nécessitent discovery, auth, session)

### Exigences fonctionnelles

- **Interopérabilité Web** : Support des clients HTTP classiques (web, mobile, tests)
- **Intégration LLM** : Exposer les capabilities comme tools pour Claude Desktop
- **Communication Agent-to-Agent** : Discovery automatique, auth sécurisée, session management

### Exigences non-fonctionnelles

- ✅ **Zero duplication** : La logique métier ne doit être implémentée qu'une fois
- ✅ **Extensibilité** : Pouvoir ajouter de nouveaux transports (GraphQL, gRPC) sans refonte
- ✅ **Conformité** : Respecter les specs UCP (REST), MCP, et A2A
- ✅ **Maintenabilité** : Chaque transport est un adaptateur pur (pas de business logic)

## Décision

Implémenter une **architecture multi-transport** avec 3 protocoles simultanés, tous déléguant à la même interface métier `merchant.Merchant`.

### Architecture en Couches

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

## Alternatives Considérées

### Alternative 1 : REST Uniquement

**Pour** :
- ✅ Simple, bien connu
- ✅ Tooling mature

**Contre** :
- ❌ LLM clients doivent parser REST
- ❌ Pas de discovery standard
- ❌ Auth custom pour agents

**Verdict** : ❌ Rejeté. Ne couvre pas besoins LLM et agents.

---

### Alternative 2 : MCP Uniquement

**Pour** :
- ✅ Tool definitions pour LLM
- ✅ JSON-RPC 2.0

**Contre** :
- ❌ Pas d'auth (local seulement)
- ❌ Sémantique LLM→Tool, pas Agent→Service
- ❌ Clients web doivent implémenter JSON-RPC

**Verdict** : ❌ Rejeté. Excellent pour LLM, pas pour web.

---

### Alternative 3 : A2A Uniquement

**Pour** :
- ✅ Discovery automatique
- ✅ OAuth2 + PKCE
- ✅ Session management

**Contre** :
- ❌ Over-engineering pour web
- ❌ LLM clients doivent implémenter OAuth2
- ❌ Complexité pour debug (curl + OAuth)

**Verdict** : ❌ Rejeté. Excellent pour agents, trop complexe pour web.

---

### Alternative 4 : Multi-Transport ✅

**Pour** :
- ✅ Chaque client utilise le transport adapté
- ✅ Zéro duplication (tous délèguent à `merchant.Merchant`)
- ✅ Extensible (pattern établi)
- ✅ Conformité specs

**Contre** :
- ❌ 3 implémentations sérialisation
- ❌ Complexité accrue
- ❌ Tests multiples

**Verdict** : ✅ **Choisi**. Avantages > inconvénients.

---

## Trade-offs

### Positifs

✅ **Flexibilité client** : Chaque client choisit son protocole  
✅ **Zero duplication** : Logique métier unique  
✅ **Validation architecture** : Ajout A2A prouve extensibilité  
✅ **Conformité multi-spec** : UCP, MCP, A2A respectés  

### Négatifs

❌ **Complexité** : 3 packages (~3300 lignes total)  
❌ **Maintenance** : 3 transports à mettre à jour  
❌ **Tests** : 60 UCP + 43 MCP + tests A2A  

---

## Tableau Comparatif

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

---

## Conséquences

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

---

## Références

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
