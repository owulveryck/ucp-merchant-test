# Agents A2A Indépendants

Agents autonomes avec communication Agent-to-Agent via JSON-RPC 2.0.

## 🎯 Agents Disponibles

### Agent Customer Growth (Port 9001)
- **Département** : Fidélisation
- **Rôle** : Analyser la valeur client et recommander des stratégies de rétention
- **Méthodes** :
  - `analyze_customer` - Analyse complète du client
  - `get_customer_tier` - Obtenir le tier du client (gold/silver/bronze)
  - `recommend_discount` - Recommander une réduction

### Agent Competitiveness (Port 9002)
- **Département** : Stratégie Prix
- **Rôle** : Analyser la compétitivité du prix et recommander une stratégie
- **Méthodes** :
  - `analyze_competitiveness` - Analyse compétitive complète
  - `check_price_position` - Vérifier position marché
  - `recommend_strategy` - Recommander stratégie pricing

## 🚀 Quick Start

### Lancer les Agents

```bash
# Terminal 1 : Agent Customer Growth
bin/customer-growth-agent --port 9001

# Terminal 2 : Agent Competitiveness  
bin/competitiveness-agent --port 9002

# Terminal 3 : Dashboard Web
bin/agents-dashboard --port 8080
```

### Utiliser le Dashboard

Ouvrir http://localhost:8080

## 📡 Utilisation CLI (curl)

### Agent Customer Growth

```bash
# Analyser un client
curl -X POST http://localhost:9001/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_customer",
    "params": {"customer_id": "elsi"},
    "id": 1
  }' | jq
```

**Réponse** :
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "agent": {
      "name": "Customer Growth Agent",
      "department": "Fidélisation",
      "role": "Analyser la valeur client..."
    },
    "message": "Bonjour, je suis Customer Growth Agent du département Fidélisation. Le client 'elsi' est un client GOLD...",
    "decision": {
      "customer_id": "elsi",
      "tier": "gold",
      "important": true,
      "suggested_discount": 10
    }
  }
}
```

### Agent Competitiveness

```bash
# Analyser compétitivité
curl -X POST http://localhost:9002/a2a \
  -H "Content-Type": application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_competitiveness",
    "params": {"product_id": "laptop", "price": 100000},
    "id": 1
  }' | jq
```

## 🔍 Endpoints des Agents

Chaque agent expose :

- `POST /a2a` - JSON-RPC 2.0 endpoint
- `GET /identity` - Carte d'identité de l'agent
- `GET /methods` - Liste des méthodes supportées
- `GET /health` - Health check

### Exemples

```bash
# Obtenir l'identité
curl http://localhost:9001/identity | jq

# Lister les méthodes
curl http://localhost:9001/methods | jq

# Health check
curl http://localhost:9001/health | jq
```

## 🧪 Données de Test

### Clients (Customer Growth)
- `elsi` - Client GOLD ($850 dépensés)
- `john` - Client SILVER ($350 dépensés)
- `alice` - Client PREMIUM ($1200 dépensés)
- `bob` - Client STANDARD ($50 dépensés)

### Produits (Competitiveness)
- `laptop` - Ordinateur portable
- `mouse` - Souris
- `keyboard` - Clavier
- `monitor` - Écran

## 🏗️ Architecture

```
┌─────────────────────────────┐
│  Dashboard Web (:8080)      │
│  Interface de test          │
└────────┬────────────────────┘
         │ HTTP
         ├─────────────┬───────────────┐
         ▼             ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌─────────────┐
│ Agent CG     │ │ Agent Comp   │ │ Client      │
│ Port: 9001   │ │ Port: 9002   │ │ Externe     │
│              │ │              │ │             │
│ JSON-RPC 2.0 │ │ JSON-RPC 2.0 │ │ (curl/...)  │
└──────────────┘ └──────────────┘ └─────────────┘
```

## 📚 Documentation

- **Framework** : [Divio Documentation](../../docs/README.md)
- **Package A2A** : `pkg/a2a/`
- **Agents source** :
  - Customer Growth : `cmd/customer-growth-agent/`
  - Competitiveness : `cmd/competitiveness-agent/`

## 🔧 Configuration

Les agents utilisent des données mock par défaut. Pour connecter à de vraies sources :

**Customer Growth** :
- Modifier `datasources.NewMockCustomerDataSource()` dans `agent.go`
- Implémenter `CustomerDataSource` interface

**Competitiveness** :
- Passer un vrai `CompetitorPriceSource` au lieu de `nil`
- Connecter à Shopping Graph ou autre source de prix

## 💡 Use Cases

### 1. Intégration dans CRM
```python
import requests

response = requests.post('http://localhost:9001/a2a', json={
    "jsonrpc": "2.0",
    "method": "analyze_customer",
    "params": {"customer_id": "elsi"},
    "id": 1
})

tier = response.json()['result']['decision']['tier']
if tier == 'gold':
    send_vip_email(customer)
```

### 2. Service Pricing
```go
// Appeler l'agent depuis Go
response, _ := http.Post("http://localhost:9002/a2a", ...)
// Parse response et utiliser
```

### 3. Webhook Automation
```bash
# Webhook sur nouveau client → analyse automatique
curl -X POST http://localhost:9001/a2a -d '...'
```

## 🚢 Déploiement

### Docker
```dockerfile
FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o agent ./cmd/customer-growth-agent
CMD ["./agent", "--port", "9001"]
```

### Cloud Run / AWS Lambda
Les agents sont stateless et peuvent être déployés en serverless.

## 🐛 Troubleshooting

**Agent ne démarre pas** :
```bash
# Vérifier port disponible
lsof -i :9001

# Lancer avec autre port
bin/customer-growth-agent --port 9003
```

**Erreur "connection refused"** :
- L'agent n'est pas lancé
- Mauvais port/URL
- Firewall bloque

**JSON-RPC error** :
- Vérifier format requête (jsonrpc: "2.0" requis)
- Méthode existe ? (`curl /methods`)
- Paramètres corrects ?

## ✅ Tests

```bash
# Test agent customer growth
curl -X POST http://localhost:9001/a2a \
  -d '{"jsonrpc":"2.0","method":"analyze_customer","params":{"customer_id":"elsi"},"id":1}' \
  | jq '.result.message'

# Doit afficher :
# "Bonjour, je suis Customer Growth Agent..."
```

---

**Créé avec** : Go 1.24 + JSON-RPC 2.0 + A2A Protocol  
**Projet** : UCP Merchant Test - Agents Autonomes
