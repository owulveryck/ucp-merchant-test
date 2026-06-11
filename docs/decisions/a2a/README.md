# ADRs - Architecture A2A (Microservices)

## 🚀 À propos

**Architecture A2A** (Agent-to-Agent) est l'implémentation **microservices** pour POCs et démos rapides.

**Composants** :
- **Customer Growth Agent** : Microservice standalone (port 9001)
- **Competitiveness Agent** : Microservice standalone (port 9002)
- **Dashboard Web** : Interface de test (port 8080)
- **Mock Data** : Données de test intégrées (pas de BDD)

**Caractéristiques** :
- ✅ Setup instantané (30 secondes)
- ✅ Aucune dépendance externe
- ✅ Agents testables en isolation
- ✅ Protocole JSON-RPC 2.0 standard

**Quand l'utiliser** :
- POC clients rapides
- Démos commerciales
- Tests d'intégration isolés
- Développement local sans infrastructure

---

## 📋 ADRs (0011-0012)

### 🤖 Agents Indépendants

**[0011 - Agents A2A Indépendants](0011-agents-a2a-independants.md)**
- **Décision** : Microservices autonomes communiquant via JSON-RPC 2.0
- **Contexte** : Simplifier les démos clients et POCs
- **Impact** : POC en 30s vs 30min de setup

### 🎮 Données de Test

**[0012 - Mock Data Sources Standalone](0012-mock-data-sources-standalone.md)**
- **Décision** : Données hardcodées dans le code
- **Contexte** : Tests sans base de données
- **Migration** : Interface unifiée mock → production
- **Impact** : Démos fonctionnent offline

---

## 🚀 Démarrage rapide

### Prérequis
- Go 1.24+
- Aucune dépendance externe (pas de Docker, BDD, Redis)

### Lancer tous les agents (1 commande)
```bash
./scripts/start-agents.sh
```

Ouvre http://localhost:8080 dans ton navigateur.

### Lancer 1 agent manuellement
```bash
# Agent Customer Growth
./bin/customer-growth-agent --port 9001

# Tester
curl -X POST http://localhost:9001/a2a \
  -d '{"jsonrpc":"2.0","method":"analyze_customer","params":{"customer_id":"elsi"},"id":1}'
```

### Arrêter tous les agents
```bash
./scripts/stop-agents.sh
```

---

## 📚 Documentation complète

**Guide Divio complet** : [../../agents-a2a-guide.md](../../agents-a2a-guide.md)
- 📚 Tutorial : Premier lancement en 5 minutes
- 🔧 How-to : Tâches pratiques (lancer, tester, configurer)
- 📖 Reference : API endpoints, clients de test, JSON-RPC
- 💡 Explanation : Concepts, architecture, décisions

**Résumé exécutif** : [../../agents-a2a-summary.md](../../agents-a2a-summary.md)
- Différence vs monolithe
- 5 arguments business clients
- Exemple concret (30s vs 30min)

**Navigation** : [../../NAVIGATION.md](../../NAVIGATION.md)
- Par objectif (tester, comprendre, intégrer, démontrer)
- Par profil (dev, commercial, architecte, chef projet)

---

## 🔗 Ressources

- **Code source** : `pkg/a2a/`, `cmd/*-agent/`
- **Scripts** : `scripts/start-agents.sh`, `scripts/stop-agents.sh`
- **Binaires** : `bin/customer-growth-agent`, `bin/competitiveness-agent`
- **Index général** : [../INDEX.md](../INDEX.md)
- **Architecture Arena** : [../arena/](../arena/) (alternative pour production)

---

## 📊 Diagramme architecture

```
┌─────────────────────────┐         ┌─────────────────────────┐
│  Customer Growth Agent  │         │  Competitiveness Agent  │
│  (Standalone)           │◄───────►│  (Standalone)           │
├─────────────────────────┤         ├─────────────────────────┤
│ • Fidélisation          │         │ • Stratégie Prix        │
│ • Port 9001             │         │ • Port 9002             │
│ • Mock customer data    │         │ • Mock competitor data  │
└───────────┬─────────────┘         └───────────┬─────────────┘
            │                                   │
            └───────────────┬───────────────────┘
                            │
                     JSON-RPC 2.0
                            │
                ┌───────────▼───────────┐
                │   Dashboard Web       │
                │   (localhost:8080)    │
                └───────────────────────┘
```

### Clients de test disponibles

| ID | Montant dépensé | Tier | Achats | Dernière activité |
|----|-----------------|------|--------|-------------------|
| `elsi` | $850 | Gold | 8 | 10 jours |
| `olwu` | $1200 | Premium | 15 | 7 jours |
| `lja` | $50 | Standard | 1 | 120 jours |
| `manu` | $350 | Silver | 4 | 20 jours |

### Produits avec concurrents

| Produit | Concurrents | Prix min | Prix max |
|---------|-------------|----------|----------|
| `laptop` | 3 | $950 | $1050 |
| `mouse` | 2 | $25 | $30 |
| `keyboard` | 3 | $68 | $75 |
| `monitor` | 2 | $350 | $380 |

---

## 💡 Exemples de requêtes

### Analyser un client
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

**Réponse** :
```json
{
  "jsonrpc": "2.0",
  "result": {
    "agent": {
      "name": "Customer Growth Agent",
      "department": "Fidélisation"
    },
    "message": "Bonjour, je suis Customer Growth Agent...",
    "decision": {
      "ShouldRetain": true,
      "CustomerTier": "gold",
      "SuggestedDiscount": 10,
      "LifetimeValue": 85000
    }
  }
}
```

### Analyser la compétitivité d'un prix
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

---

## 🆚 A2A vs Arena

| Critère | A2A (Microservices) | Arena (Monolithe) |
|---------|---------------------|-------------------|
| **Setup** | 30 secondes | 30 minutes |
| **Dépendances** | Aucune | Docker, Postgres, Redis |
| **Données** | Mock intégré | BDD réelle |
| **Performance** | Bonne | Maximum |
| **Déploiement** | Par agent | Global |
| **Usage** | POC, Démos, Tests | Production |

**Principe** : Les deux architectures **partagent la même logique métier** (agents pricing), seule la couche transport diffère.

---

*Architecture microservices pour POCs et démos rapides*
