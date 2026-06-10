# ADRs - Architecture Arena (Monolithe)

## 🏛️ À propos

**Architecture Arena** est l'implémentation **monolithique** multi-agents pour la production.

**Composants** :
- **Shopping Graph** : Orchestrateur du parcours d'achat
- **Observability Hub** : Monitoring centralisé
- **Arena** : Environnement d'exécution des agents
- **Agents pricing** : CustomerGrowth, Competitiveness, etc.

**Caractéristiques** :
- ✅ Performance maximale (communication in-process)
- ✅ Transactions ACID (cohérence forte)
- ✅ Base de données réelle (Postgres/Redis)
- ✅ Production-ready

**Quand l'utiliser** :
- Déploiement production
- Performance critique
- Transactions distribuées complexes
- Forte cohérence de données nécessaire

---

## 📋 ADRs (0001-0010)

### 🤖 Multi-Agent & Orchestration

- **[0001](0001-architecture-multi-agents-pour-prix-competitif.md)** - Architecture multi-agents pour prix compétitif
- **[0004](0004-architecture-3-agents-orchestree.md)** - Architecture 3 agents orchestrée
- **[0005](0005-agent-acheteur-integre.md)** - Agent acheteur intégré
- **[0008](0008-multi-agent-shopping-architecture.md)** - Multi-agent shopping architecture
- **[0010](0010-competitive-pricing-agent.md)** - Competitive pricing agent

### 💰 Stratégies de Pricing

- **[0002](0002-strategie-victoire-avant-marge-parfaite.md)** - Stratégie victoire avant marge parfaite
- **[0003](0003-strategie-detection-codes-promo.md)** - Stratégie détection codes promo

### 📡 Transport & Communication

- **[0009](0009-multi-transport-architecture.md)** - Multi-transport architecture (REST + MCP)

### 🎯 Observabilité

- **[0006](0006-messages-detailles-decision-achat.md)** - Messages détaillés décision achat

### 🎮 Tests

- **[0007](0007-scenario-challenge-concurrents.md)** - Scénario challenge concurrents

---

## 🚀 Démarrage rapide

### Prérequis
- Docker + Docker Compose
- Base de données Postgres
- Redis (optionnel, pour cache)

### Lancer Arena
```bash
cd demo/
docker-compose up -d postgres redis
go run cmd/arena/main.go
```

### Lancer Shopping Graph + Obs Hub
```bash
# Terminal 1
go run cmd/shopping-graph/main.go --port 8081

# Terminal 2
go run cmd/obs-hub/main.go --port 8082
```

### Tester
```bash
# Via REST API
curl -X POST http://localhost:8081/api/purchase \
  -d '{"product_id": "laptop", "customer_id": "alice"}'

# Via MCP (JSON-RPC)
curl -X POST http://localhost:8081/mcp \
  -d '{"jsonrpc":"2.0","method":"analyze_price","params":{"product":"laptop"},"id":1}'
```

---

## 🔗 Ressources

- **Code source** : `demo/`, `pkg/merchant/`
- **Tests** : `scripts/arena_challenge.sh`
- **Index général** : [../INDEX.md](../INDEX.md)
- **Architecture A2A** : [../a2a/](../a2a/) (alternative pour POCs)

---

## 📊 Diagramme architecture

```
┌─────────────────────────────────────────────────────┐
│                  Shopping Graph                     │
│  (Orchestrateur parcours client)                    │
└────────────┬────────────────────────────────────────┘
             │
    ┌────────┴────────┬──────────────┐
    │                 │              │
┌───▼────┐      ┌────▼─────┐   ┌───▼──────┐
│Customer│      │Competi-  │   │ Other    │
│Growth  │      │tiveness  │   │ Agents   │
│Agent   │      │Agent     │   │          │
└───┬────┘      └────┬─────┘   └───┬──────┘
    │                │              │
    └────────┬───────┴──────────────┘
             │
    ┌────────▼─────────────┐
    │  Observability Hub   │
    │  (Monitoring)        │
    └──────────────────────┘
             │
    ┌────────▼─────────────┐
    │    Database          │
    │  (Postgres + Redis)  │
    └──────────────────────┘
```

---

*Architecture monolithique pour production haute performance*
