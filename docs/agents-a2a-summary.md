# Agents A2A - Système Multi-Agent Autonome

## 🎯 Objectif
Créer des agents intelligents **indépendants** qui communiquent entre eux via un protocole standard

## 🏗️ Architecture

```
┌─────────────────────┐         ┌─────────────────────┐
│  Customer Growth    │         │  Competitiveness    │
│  Agent              │◄───────►│  Agent              │
├─────────────────────┤         ├─────────────────────┤
│ • Fidélisation      │         │ • Stratégie Prix    │
│ • Tiers clients     │         │ • Analyse marché    │
│ • Recommandations   │         │ • 4 sous-agents     │
└─────────────────────┘         └─────────────────────┘
         │                               │
         └───────────┬───────────────────┘
                     │
              JSON-RPC 2.0
                     │
         ┌───────────▼───────────┐
         │   Dashboard Web       │
         │   localhost:8080      │
         └───────────────────────┘
```

## ✨ Caractéristiques

**Carte d'identité des agents**
- Nom, Département, Rôle, Version
- Réponses conversationnelles en français
- Endpoints standards : `/a2a`, `/identity`, `/methods`, `/health`

**Standalone & Mock Data**
- Fonctionnent sans infrastructure Arena
- Sources de données mock intégrées
- Déployables indépendamment (ports 9001, 9002)

**Protocole JSON-RPC 2.0**
- Standard de communication inter-agents
- Request/Response structurés
- Méthodes discoverables

## 🚀 Cas d'usage

**Customer Growth** : `analyze_customer(customer_id)` → Tier, LTV, Discount
**Competitiveness** : `analyze_competitiveness(product_id, price)` → Stratégie, Prix recommandé

## 🎁 Résultat
2 microservices autonomes + Dashboard interactif = **Système multi-agent testable et démontrable**
