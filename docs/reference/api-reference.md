# Reference : API Complete

## Base URLs

| Transport | URL | Description |
|-----------|-----|-------------|
| REST | `http://localhost:8182/shopping-api` | UCP Shopping Service |
| MCP | `http://localhost:8182/mcp` | JSON-RPC 2.0 |
| A2A | `http://localhost:8182/a2a` | Agent-to-Agent |
| Discovery | `http://localhost:8182/.well-known/ucp` | UCP Discovery |
| Dashboard | `http://localhost:8182/` | SSE Dashboard |

## UCP Discovery

### GET `/.well-known/ucp`

**Response** :
```json
{
  "version": "2026-01-11",
  "endpoints": {
    "shopping": "http://localhost:8182/shopping-api"
  },
  "capabilities": []
}
```

## REST API - Checkout Sessions

### POST `/shopping-api/checkout-sessions`

Créer une nouvelle checkout session.

**Request** :
```json
{
  "items": [
    {
      "item_id": "rose_bouquet",
      "quantity": 2
    }
  ],
  "buyer": {
    "email": "john.doe@example.com"
  }
}
```

**Response** `201 Created` :
```json
{
  "id": "checkout_abc123",
  "status": "incomplete",
  "ucp": {
    "version": "2026-01-11",
    "capabilities": []
  },
  "items": [...],
  "totals": [...],
  "payment": {
    "methods": [...]
  },
  "fulfillment": {
    "methods": [...]
  }
}
```

### GET `/shopping-api/checkout-sessions/{id}`

Récupérer une checkout session.

**Response** `200 OK` : même structure que POST.

### PUT `/shopping-api/checkout-sessions/{id}`

Mettre à jour une checkout session.

**Request** :
```json
{
  "discount_codes": ["WELCOME20"],
  "fulfillment": {
    "destination": {
      "country": "US",
      "postal_code": "10001"
    },
    "selection": {
      "method_id": "shipping",
      "destination_id": "dest_001",
      "group_id": "group_standard",
      "option_id": "standard_5_7"
    }
  },
  "payment": {
    "selection": {
      "method_id": "card"
    },
    "card_credential": {
      "token": "success_token",
      "handler_id": "stripe_test"
    }
  }
}
```

**Response** `200 OK` : checkout session mise à jour.

### POST `/shopping-api/checkout-sessions/{id}/complete`

Finaliser le checkout et créer une commande.

**Headers** :
```
Idempotency-Key: unique-key-123
```

**Request** :
```json
{
  "buyer_consent": {
    "marketing": true,
    "analytics": false,
    "sale_of_data": false
  }
}
```

**Response** `201 Created` :
```json
{
  "order": {
    "id": "ord_xyz789",
    "status": "pending",
    "ucp": {
      "version": "2026-01-11",
      "capabilities": []
    },
    "items": [...],
    "totals": [...],
    "fulfillment": {...},
    "payment": {...}
  },
  "links": [
    {
      "href": "http://localhost:8182/orders/ord_xyz789",
      "type": "application/json"
    }
  ]
}
```

### POST `/shopping-api/checkout-sessions/{id}/cancel`

Annuler une checkout session.

**Response** `200 OK` :
```json
{
  "id": "checkout_abc123",
  "status": "canceled"
}
```

## REST API - Orders

### GET `/orders/{id}`

Récupérer une commande.

**Response** `200 OK` :
```json
{
  "id": "ord_xyz789",
  "status": "pending",
  "ucp": {
    "version": "2026-01-11",
    "capabilities": []
  },
  "items": [...],
  "totals": [...],
  "fulfillment": {
    "status": "pending",
    "shipments": []
  },
  "payment": {
    "status": "authorized"
  }
}
```

### PUT `/orders/{id}`

Mettre à jour une commande (webhook URL).

**Request** :
```json
{
  "webhook_url": "https://example.com/webhooks/ucp"
}
```

## Simulation Endpoint

### POST `/testing/simulate-shipping/{order_id}`

Simuler l'expédition d'une commande (tests uniquement).

**Headers** :
```
Simulation-Secret: super-secret-sim-key
```

**Response** `200 OK` :
```json
{
  "id": "ord_xyz789",
  "fulfillment": {
    "status": "shipped",
    "shipments": [
      {
        "tracking_number": "TRACK123",
        "carrier": "USPS",
        "estimated_delivery": "2026-06-15"
      }
    ]
  }
}
```

## MCP API (JSON-RPC 2.0)

### Endpoint : POST `/mcp`

**Headers** :
```
Content-Type: application/json
```

### Tool : `checkout_create`

**Request** :
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "checkout_create",
    "arguments": {
      "items": [
        {"item_id": "rose_bouquet", "quantity": 2}
      ]
    }
  }
}
```

**Response** :
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"id\": \"checkout_abc\", \"status\": \"incomplete\", ...}"
      }
    ]
  }
}
```

### Tool : `catalog_search`

**Request** :
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "catalog_search",
    "arguments": {
      "query": "roses"
    }
  }
}
```

## A2A API

### POST `/a2a`

Agent-to-Agent communication (même format JSON-RPC que MCP).

**Tools disponibles** :
- `catalog_search` - Recherche produits
- `get_pricing` - Obtenir prix avec agents
- `checkout_create` - Créer checkout
- `checkout_update` - Mettre à jour checkout

## Shopping Graph API

### GET `/search?q={query}`

Recherche cross-merchant.

**Request** :
```
GET http://localhost:9000/search?q=headphones
```

**Response** :
```json
{
  "results": [
    {
      "merchant_id": "merchant_a",
      "merchant_name": "SuperShop",
      "product_id": "wireless_headphones",
      "title": "Wireless Headphones Premium",
      "price": 8999,
      "discount_hints": ["SAVE10", "WELCOME15"],
      "in_stock": true
    }
  ]
}
```

### POST `/admin/refresh`

Forcer un refresh des catalogues merchants.

**Response** `200 OK`.

## Observability Hub API

### GET `/events`

Stream SSE des events temps réel.

**Response** (Server-Sent Events) :
```
event: vendor_decision
data: {"merchant":"MegaStore","product":"casque_audio","price":5230}

event: price_update
data: {"merchant":"MegaStore","product":"casque_audio","old":6215,"new":5230}
```

## Error Responses

Toutes les erreurs suivent le format :

```json
{
  "detail": "Error message here"
}
```

### Status Codes

| Code | Meaning | Exemple |
|------|---------|---------|
| 400 | Bad Request | Paramètre invalide |
| 402 | Payment Required | Paiement refusé |
| 404 | Not Found | Checkout/Order introuvable |
| 409 | Conflict | Idempotency key mismatch |
| 422 | Unprocessable Entity | Validation échouée |
| 500 | Internal Server Error | Erreur serveur |

## Types de Totals

Les totals valides sont :

- `items_discount` - Remise sur items (>= 0)
- `subtotal` - Sous-total avant remises
- `discount` - Remise totale appliquée (>= 0)
- `fulfillment` - Frais de livraison
- `tax` - Taxes
- `fee` - Frais additionnels
- `total` - Total final

**Important** : Les montants de discount doivent être **positifs** (valeur absolue).

## Versioning

- Header `UCP-Agent` pour négociation de version
- Version actuelle : `2026-01-11`
- Format : `"application/json; version=2026-01-11"`

## Rate Limiting

Aucune limite actuellement (test server).

En production, recommandé :
- 100 req/min par IP pour REST
- 1000 req/min pour MCP tools
