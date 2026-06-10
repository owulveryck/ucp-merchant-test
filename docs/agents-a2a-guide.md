# Guide Agents A2A - Documentation Divio

## 📚 Tutorial - Premier lancement (apprentissage)

### Objectif
Lancer votre premier agent A2A et faire votre première requête en 5 minutes.

### Prérequis
- Go 1.24+ installé
- Terminal ouvert
- Répertoire : `~/stageocto/ucp-merchant-test`

### Étape 1 : Compiler les agents
```bash
cd ~/stageocto/ucp-merchant-test
go build -o bin/customer-growth-agent ./cmd/customer-growth-agent
go build -o bin/competitiveness-agent ./cmd/competitiveness-agent
go build -o bin/agents-dashboard ./cmd/agents-dashboard
```

### Étape 2 : Lancer un agent (Customer Growth)
```bash
./bin/customer-growth-agent --port 9001
```

Vous devriez voir :
```
[Customer Growth Agent] Starting A2A server on :9001
[Customer Growth Agent] Department: Fidélisation
[Customer Growth Agent] Endpoints:
  - POST   :9001/a2a       (JSON-RPC 2.0)
  - GET    :9001/identity  (Agent identity)
```

### Étape 3 : Tester l'agent (nouveau terminal)
```bash
curl -X POST http://localhost:9001/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_customer",
    "params": {"customer_id": "elsi"},
    "id": 1
  }'
```

### Résultat attendu
```json
{
  "jsonrpc": "2.0",
  "result": {
    "agent": {
      "name": "Customer Growth Agent",
      "department": "Fidélisation",
      "role": "Analyser la valeur client..."
    },
    "message": "Bonjour, je suis Customer Growth Agent du département Fidélisation. Le client 'elsi' est un client gold ayant dépensé $850.00. OUI, c'est un client important à conserver...",
    "decision": {
      "ShouldRetain": true,
      "CustomerTier": "gold",
      "SuggestedDiscount": 10,
      "LifetimeValue": 85000
    }
  },
  "id": 1
}
```

### Étape 4 : Lancer le dashboard (optionnel)
```bash
# Terminal 3
./bin/agents-dashboard --port 8080
```

Ouvrir http://localhost:8080 dans votre navigateur.

**✅ Félicitations !** Vous avez lancé votre premier agent A2A autonome.

---

## 🔧 How-to Guides - Tâches pratiques

### Comment lancer tous les agents en une commande
```bash
./scripts/start-agents.sh
```

Vérifie que les 3 services sont lancés :
- Customer Growth Agent → port 9001
- Competitiveness Agent → port 9002  
- Dashboard → port 8080

### Comment arrêter tous les agents
```bash
./scripts/stop-agents.sh
```

### Comment tester l'agent de compétitivité
```bash
curl -X POST http://localhost:9002/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_competitiveness",
    "params": {
      "product_id": "laptop",
      "price": 100000
    },
    "id": 1
  }'
```

### Comment découvrir les méthodes d'un agent
```bash
# Liste les méthodes disponibles
curl http://localhost:9001/methods

# Obtient l'identité de l'agent
curl http://localhost:9001/identity

# Vérifie la santé de l'agent
curl http://localhost:9001/health
```

### Comment ajouter un nouveau client de test
Éditer `pkg/pricing-unified/datasources/mock_customer_data.go` :

```go
"nouveau_client": {
    CustomerID:       "nouveau_client",
    TotalSpent:       50000,  // $500
    PurchaseCount:    3,
    LastPurchaseDays: 5,
},
```

Recompiler l'agent :
```bash
go build -o bin/customer-growth-agent ./cmd/customer-growth-agent
./scripts/stop-agents.sh
./scripts/start-agents.sh
```

### Comment changer le port d'un agent
```bash
./bin/customer-growth-agent --port 9999
```

---

## 📖 Reference - Documentation technique

### Endpoints disponibles

#### Customer Growth Agent (port 9001)

**POST /a2a** - JSON-RPC 2.0
- Méthode : `analyze_customer`
  - Params : `{"customer_id": "string"}`
  - Returns : `{ShouldRetain, CustomerTier, SuggestedDiscount, LifetimeValue, RetentionReasoning}`

- Méthode : `get_customer_tier`
  - Params : `{"customer_id": "string"}`
  - Returns : `{tier: "standard"|"silver"|"gold"|"premium"}`

- Méthode : `recommend_discount`
  - Params : `{"customer_id": "string"}`
  - Returns : `{discount_percent: number, reasoning: string}`

**GET /identity**
```json
{
  "name": "Customer Growth Agent",
  "department": "Fidélisation",
  "role": "Analyser la valeur client et recommander des stratégies de rétention",
  "version": "1.0.0"
}
```

**GET /methods**
```json
["analyze_customer", "get_customer_tier", "recommend_discount"]
```

**GET /health**
```json
{"status": "ok"}
```

#### Competitiveness Agent (port 9002)

**POST /a2a** - JSON-RPC 2.0
- Méthode : `analyze_competitiveness`
  - Params : `{"product_id": "string", "price": number}`
  - Returns : `{IsCompetitive, MarketPosition, Strategy, RecommendedPrice, Margin, ...}`

- Méthode : `check_price_position`
  - Params : `{"product_id": "string", "price": number}`
  - Returns : `{market_position: number, total_competitors: number, is_competitive: bool}`

- Méthode : `recommend_strategy`
  - Params : `{"product_id": "string", "price": number}`
  - Returns : `{strategy: string, recommended_price: number, margin: number}`

**GET /identity**, **GET /methods**, **GET /health** - Identique à Customer Growth

#### Dashboard (port 8080)

**GET /** - Interface web interactive

**GET /api/agents** - Liste des agents disponibles
```json
[
  {"name": "Customer Growth", "url": "http://localhost:9001", "port": 9001},
  {"name": "Competitiveness", "url": "http://localhost:9002", "port": 9002}
]
```

**POST /api/call** - Proxy vers un agent
- Body : `{"agent_url": "http://localhost:9001", "request": {...}}`

### Clients de test disponibles

| ID | Total dépensé | Tier | Achats | Dernière activité |
|----|---------------|------|--------|-------------------|
| `elsi` | $850 | Gold | 8 | 10 jours |
| `alice` | $1200 | Premium | 15 | 7 jours |
| `bob` | $50 | Standard | 1 | 120 jours |
| `john` | $350 | Silver | 4 | 20 jours |

### Produits avec données concurrents

| Produit | Concurrents | Prix min | Prix max |
|---------|-------------|----------|----------|
| `laptop` | 3 | $950 | $1050 |
| `mouse` | 2 | $25 | $30 |
| `keyboard` | 3 | $68 | $75 |
| `monitor` | 2 | $350 | $380 |

### Structure JSON-RPC 2.0

**Requête**
```json
{
  "jsonrpc": "2.0",
  "method": "nom_methode",
  "params": {"param1": "value1"},
  "id": 1
}
```

**Réponse succès**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "agent": {...},
    "message": "...",
    "decision": {...}
  },
  "id": 1
}
```

**Réponse erreur**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "Method not found"
  },
  "id": 1
}
```

### Ports par défaut

| Service | Port | Modifiable via |
|---------|------|----------------|
| Customer Growth | 9001 | `--port 9001` |
| Competitiveness | 9002 | `--port 9002` |
| Dashboard | 8080 | `--port 8080` |

---

## 💡 Explanation - Comprendre les concepts

### Pourquoi des agents indépendants ?

#### Problème avec l'architecture monolithique (Arena)

**Couplage fort**
- Tous les agents vivent dans un seul processus
- Impossible de déployer un agent sans les autres
- Une erreur dans un agent peut crasher toute l'application

**Dépendances complexes**
- Shopping Graph (gestion du parcours client)
- Observability Hub (monitoring)
- Arena (orchestration)
- Base de données partagée

**Démonstration difficile**
- Setup complet nécessaire même pour une simple démo
- 30+ minutes de configuration
- Risque d'échec technique pendant la démo client

#### Solution : Architecture A2A (Agent-to-Agent)

**Microservices autonomes**
Chaque agent est un binaire indépendant :
- ✅ Démarrage : 1 commande
- ✅ Dépendances : 0 (données mock intégrées)
- ✅ Mémoire : ~10 MB par agent
- ✅ Crash isolation : 1 agent down ≠ système down

**Protocole standard JSON-RPC 2.0**
- Standard IETF (RFC 4627)
- Compatible tous langages (Go, Python, JavaScript, Java...)
- Transport agnostique (HTTP, WebSocket, TCP...)
- Discovery built-in (`/methods`, `/identity`)

**Communication inter-agents**
```
Client → Agent A → Agent B → Agent C
         (JSON-RPC)  (JSON-RPC)
```

Au lieu de :
```
Client → Monolithe [Agent A + Agent B + Agent C]
```

### Quand utiliser l'architecture A2A ?

**✅ Utilisez A2A pour :**
- POC (Proof of Concept) rapides
- Démos clients
- Tests d'intégration isolés
- Environnements avec contraintes mémoire
- Déploiements progressifs (1 agent à la fois)

**❌ Préférez le monolithe pour :**
- Production à très haute performance (latence inter-agents)
- Transactions distribuées complexes
- Environnements où le réseau est instable

### Comment les agents communiquent-ils ?

**Exemple : Calcul de prix avec 2 agents**

1. **Client** envoie une requête au **Customer Growth Agent**
```json
POST http://localhost:9001/a2a
{"method": "analyze_customer", "params": {"customer_id": "elsi"}}
```

2. **Customer Growth Agent** analyse et répond
```json
{"result": {"tier": "gold", "discount": 10}}
```

3. Le client utilise cette info pour interroger **Competitiveness Agent**
```json
POST http://localhost:9002/a2a
{"method": "analyze_competitiveness", "params": {"product_id": "laptop", "price": 100000}}
```

4. **Competitiveness Agent** calcule le prix optimal
```json
{"result": {"recommended_price": 99000, "strategy": "match_lowest"}}
```

**Résultat** : Prix final = $990 avec 10% de réduction pour client Gold

### Mock Data vs Production

**Mock Data (agents A2A)**
- Données hardcodées dans le code
- 4 clients de test prédéfinis
- Pas de base de données nécessaire
- Reproductible à 100%

**Production (Arena)**
- Connexion à une vraie base de données
- Clients réels dynamiques
- APIs externes pour prix concurrents
- Cache Redis pour performance

**Migration Mock → Production**
Il suffit de remplacer :
```go
// Mock
dataSource := datasources.NewMockCustomerDataSource()

// Production
dataSource := datasources.NewPostgresCustomerDataSource(dbConnection)
```

L'interface reste identique → **aucun changement dans la logique métier**.

### Avantages du protocole JSON-RPC 2.0

**1. Simplicité**
- 4 champs seulement : `jsonrpc`, `method`, `params`, `id`
- Facile à debugger (JSON lisible)
- Testable avec `curl`

**2. Découvrabilité**
- `GET /methods` → liste des méthodes disponibles
- `GET /identity` → qui est cet agent ?
- Pas besoin de documentation externe

**3. Versioning naturel**
```json
{"method": "analyze_customer_v2", "params": {...}}
```

**4. Batch requests** (future)
```json
[
  {"method": "analyze_customer", "params": {...}, "id": 1},
  {"method": "recommend_discount", "params": {...}, "id": 2}
]
```

### Évolution future

**Étape actuelle** : Agents standalone avec mock data

**Prochaines étapes** :
1. **Service Discovery** : Les agents se trouvent automatiquement
2. **Load Balancing** : Plusieurs instances du même agent
3. **Authentification** : JWT tokens pour sécuriser les appels
4. **Streaming** : WebSocket pour mises à jour temps réel
5. **Orchestration** : Agent coordinator qui route les requêtes

**Vision long terme** : Marketplace d'agents
- Agents vendus individuellement
- Clients composent leur propre plateforme
- Pay-per-use par agent

---

## 🎯 Résumé par profil

### Pour un **Développeur**
- Lancez `./scripts/start-agents.sh`
- Testez avec `curl` ou le dashboard
- Ajoutez des clients dans `mock_customer_data.go`
- Créez de nouveaux agents en copiant la structure existante

### Pour un **Commercial**
- Démo en 30 secondes : `./bin/customer-growth-agent --port 9001` + 1 requête curl
- Dashboard visuel pour clients non techniques
- Arguments business : coûts réduits, déploiement rapide, pas de lock-in

### Pour un **Architecte**
- Microservices Go indépendants
- JSON-RPC 2.0 standard
- Mock data pour tests, interfaces pour production
- Scalable horizontalement

### Pour un **Chef de projet**
- POC livrable en jours vs semaines
- Démos clients sans risque technique
- Déploiement progressif (1 agent → plateforme complète)
- ROI rapide grâce à la simplicité
