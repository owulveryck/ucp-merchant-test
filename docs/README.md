# Documentation UCP Merchant Test

Cette documentation suit le [framework Divio](https://documentation.divio.com/) pour une organisation claire et efficace.

## Structure de la documentation

### 📚 [Tutorials](tutorials/) - Apprendre en pratiquant
*Learning-oriented* : Guides pas-à-pas pour débutants qui veulent découvrir le projet.

- [Getting Started](tutorials/01-getting-started.md) - Première prise en main
- [Running Your First Demo](tutorials/02-first-demo.md) - Lancer la démo shopping
- [Understanding Multi-Agent Pricing](tutorials/03-multi-agent-pricing.md) - Comprendre le pricing intelligent

### 🔧 [How-to Guides](how-to/) - Accomplir des tâches spécifiques
*Task-oriented* : Solutions pratiques pour résoudre des problèmes concrets.

- [Configure a New Merchant](how-to/configure-merchant.md) - Ajouter un marchand
- [Set Up Custom Pricing Strategies](how-to/setup-pricing-strategies.md) - Configurer les stratégies
- [Run Conformance Tests](how-to/run-conformance-tests.md) - Exécuter les tests UCP
- [Monitor Agent Decisions](how-to/monitor-agent-decisions.md) - Observer les agents en temps réel

### 📖 [Reference](reference/) - Informations techniques détaillées
*Information-oriented* : Documentation technique précise et exhaustive.

- [API Reference](reference/api-reference.md) - Endpoints REST/MCP/A2A
- [Configuration Options](reference/configuration.md) - Flags et variables d'environnement
- [Agent System Architecture](reference/agent-architecture.md) - Architecture 3-agents
- [Data Models](reference/data-models.md) - Structures UCP
- [Error Codes](reference/error-codes.md) - Codes d'erreur et sentinels

### 💡 [Explanation](explanation/) - Comprendre les concepts
*Understanding-oriented* : Contexte, design rationale, et vision globale.

- [Why Multi-Agent Pricing](explanation/why-multi-agent.md) - Raison d'être du système
- [UCP Protocol Integration](explanation/ucp-integration.md) - Pourquoi UCP/MCP/A2A
- [Competitive Strategy Trade-offs](explanation/competitive-tradeoffs.md) - Compromis marge vs victoire
- [System Design Philosophy](explanation/design-philosophy.md) - Principes d'architecture

### 📋 [Architecture Decision Records](decisions/)
*Context-oriented* : Historique des décisions d'architecture (ADRs).

Voir [decisions/README.md](decisions/README.md) pour l'index complet des 10 ADRs.

## Navigation rapide

**Je débute sur le projet** → Commencez par [tutorials/01-getting-started.md](tutorials/01-getting-started.md)

**Je veux faire quelque chose de précis** → Consultez [how-to/](how-to/)

**Je cherche une information technique** → Consultez [reference/](reference/)

**Je veux comprendre pourquoi c'est fait ainsi** → Consultez [explanation/](explanation/) et [decisions/](decisions/)

## Principes du framework Divio

| Type | Orientation | Analogie | Objectif |
|------|-------------|----------|----------|
| **Tutorial** | Learning | Cours de cuisine | Apprendre en pratiquant |
| **How-to** | Task | Recette | Résoudre un problème |
| **Reference** | Information | Encyclopédie | Trouver une info précise |
| **Explanation** | Understanding | Article scientifique | Comprendre en profondeur |

## Contribuer à la documentation

Lors de l'ajout de documentation, demandez-vous :
- **Est-ce pour apprendre ?** → Tutorial
- **Est-ce pour faire quelque chose ?** → How-to
- **Est-ce une info technique ?** → Reference
- **Est-ce pour comprendre le pourquoi ?** → Explanation
- **Est-ce une décision d'architecture ?** → ADR dans decisions/

## Liens externes

- [Divio Documentation System](https://documentation.divio.com/)
- [UCP Specification](https://ucp.dev)
- [MCP Protocol](https://modelcontextprotocol.io/)
