# Index des ADRs - UCP Merchant Test

## 🏛️ Deux architectures complémentaires

Ce projet implémente **deux architectures** pour répondre à des besoins différents :

### Architecture Arena (Monolithe) → `arena/`
**Pour** : Production, performance max, transactions complexes  
**ADRs** : 0001-0010  
**Code** : `demo/`, orchestration Shopping Graph + Obs Hub

### Architecture A2A (Microservices) → `a2a/`
**Pour** : POC rapides, démos clients, tests isolés  
**ADRs** : 0011-0012  
**Code** : `pkg/a2a/`, `cmd/*-agent/`, standalone binaries

---

## 🏛️ Architecture Arena (Monolithe)

**Répertoire** : [`arena/`](arena/)

### 🤖 Multi-Agent & Orchestration

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0001](arena/0001-architecture-multi-agents-pour-prix-competitif.md) | Architecture multi-agents pour prix compétitif | Décision d'utiliser plusieurs agents spécialisés | 2024-05-29 |
| [0004](arena/0004-architecture-3-agents-orchestree.md) | Architecture 3 agents orchestrée | Price Intelligence, Market Analysis, Strategy Recommender | 2024-06-04 |
| [0005](arena/0005-agent-acheteur-integre.md) | Agent acheteur intégré | Shopping Agent dans l'architecture Arena | 2024-06-04 |
| [0008](arena/0008-multi-agent-shopping-architecture.md) | Multi-agent shopping architecture | Architecture complète Shopping Graph + Obs Hub | 2024-05-26 |
| [0010](arena/0010-competitive-pricing-agent.md) | Competitive pricing agent | Agent de pricing compétitif 4-sous-agents | 2024-05-28 |

### 💰 Stratégies de Pricing

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0002](arena/0002-strategie-victoire-avant-marge-parfaite.md) | Stratégie victoire avant marge parfaite | Prioriser la compétitivité vs marge max | 2024-05-29 |
| [0003](arena/0003-strategie-detection-codes-promo.md) | Stratégie détection codes promo | Recherche intelligente de codes promotionnels | 2024-05-29 |

### 📡 Transport & Communication

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0009](arena/0009-multi-transport-architecture.md) | Multi-transport architecture | REST + MCP (JSON-RPC) + A2A | 2024-06-04 |

### 🎯 Observabilité & Messages

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0006](arena/0006-messages-detailles-decision-achat.md) | Messages détaillés décision achat | Traçabilité des décisions agents | 2024-06-04 |

### 🎮 Scénarios & Tests

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0007](arena/0007-scenario-challenge-concurrents.md) | Scénario challenge concurrents | Environnement de test multi-agents | 2024-06-04 |

---

## 🚀 Architecture A2A (Microservices)

**Répertoire** : [`a2a/`](a2a/)

### 🤖 Agents Indépendants

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0011](a2a/0011-agents-a2a-independants.md) | **Agents A2A Indépendants** | Microservices autonomes JSON-RPC 2.0 | 2024-06-09 |

### 🎮 Données de Test

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0012](a2a/0012-mock-data-sources-standalone.md) | **Mock Data Sources standalone** | Données de test intégrées (4 clients, 4 produits) | 2024-06-09 |

---

## 🔄 Par chronologie

### Mai 2024 - Architecture Arena

```
2024-05-26
  └── arena/0008 - Multi-agent shopping architecture

2024-05-28
  └── arena/0010 - Competitive pricing agent

2024-05-29
  ├── arena/0001 - Architecture multi-agents pour prix compétitif
  ├── arena/0002 - Stratégie victoire avant marge parfaite
  └── arena/0003 - Stratégie détection codes promo
```

### Juin 2024 - Évolutions Arena + Nouveauté A2A

```
2024-06-04 (Arena)
  ├── arena/0004 - Architecture 3 agents orchestrée
  ├── arena/0005 - Agent acheteur intégré
  ├── arena/0006 - Messages détaillés décision achat
  ├── arena/0007 - Scénario challenge concurrents
  └── arena/0009 - Multi-transport architecture

2024-06-09 (A2A 🆕 Nouvelle architecture)
  ├── a2a/0011 - Agents A2A Indépendants
  └── a2a/0012 - Mock Data Sources standalone
```

---

## 🎯 Par cas d'usage

### Je veux faire une démo rapide (POC client)
➜ **Architecture A2A**
1. [ADR-0011: Agents A2A Indépendants](a2a/0011-agents-a2a-independants.md) ⭐
2. [ADR-0012: Mock Data Sources standalone](a2a/0012-mock-data-sources-standalone.md)

### Je veux déployer en production
➜ **Architecture Arena**
1. [ADR-0008: Multi-agent shopping architecture](arena/0008-multi-agent-shopping-architecture.md)
2. [ADR-0009: Multi-transport architecture](arena/0009-multi-transport-architecture.md)
3. [ADR-0010: Competitive pricing agent](arena/0010-competitive-pricing-agent.md)

### Je veux comprendre le pricing compétitif
➜ **Architecture Arena**
1. [ADR-0001: Architecture multi-agents pour prix compétitif](arena/0001-architecture-multi-agents-pour-prix-competitif.md)
2. [ADR-0004: Architecture 3 agents orchestrée](arena/0004-architecture-3-agents-orchestree.md)
3. [ADR-0010: Competitive pricing agent](arena/0010-competitive-pricing-agent.md)

### Je veux comprendre les stratégies business
➜ **Architecture Arena**
1. [ADR-0002: Stratégie victoire avant marge parfaite](arena/0002-strategie-victoire-avant-marge-parfaite.md)
2. [ADR-0003: Stratégie détection codes promo](arena/0003-strategie-detection-codes-promo.md)

---

## 📊 Statistiques

- **Total ADRs** : 12
  - **Arena** : 10 ADRs (0001-0010)
  - **A2A** : 2 ADRs (0011-0012)

**Par thématique** :
- Architecture : 6 ADRs
- Stratégies : 2 ADRs
- Tests & Données : 2 ADRs
- Observabilité : 1 ADR
- Transport : 1 ADR

---

## 🆕 Nouveauté : Architecture A2A (Juin 2024)

### 🚀 Nouvelle branche d'architecture

**Constat** : L'architecture Arena est excellente pour la production mais trop lourde pour des POCs rapides.

**Solution** : Création d'une **architecture parallèle A2A** (Agent-to-Agent) basée sur microservices.

### ADR-0011: Agents A2A Indépendants
**Changement** : Ajout d'une architecture microservices **en complément** d'Arena (pas en remplacement).

| Critère | Arena (Monolithe) | A2A (Microservices) |
|---------|-------------------|---------------------|
| **Usage** | Production | POC, Démos, Tests |
| **Setup** | 30+ minutes | 30 secondes |
| **Déploiement** | Global | Sélectif (1 agent) |
| **Couplage** | Fort | Zéro |
| **Performance** | Maximum | Bonne |
| **Données** | BDD réelle | Mock intégré |

**Impact business** :
- ✅ POC livrables en jours vs semaines
- ✅ Démos clients sans risque technique
- ✅ Coûts infrastructure réduits de 90%

**Lien** : [ADR-0011 complet](a2a/0011-agents-a2a-independants.md)

### ADR-0012: Mock Data Sources
**Complément** : Données de test intégrées permettant aux agents A2A de fonctionner sans BDD.

**Données fournies** :
- 4 clients de test : elsi ($850), alice ($1200), bob ($50), john ($350)
- 4 produits avec concurrents : laptop, mouse, keyboard, monitor
- Zéro configuration nécessaire

**Impact développement** :
- ✅ Tests 100% reproductibles
- ✅ Démos fonctionnent offline (train, avion, hôtel)
- ✅ Onboarding développeurs instantané (git clone → go run)

**Lien** : [ADR-0012 complet](a2a/0012-mock-data-sources-standalone.md)

---

### 🤝 Les deux architectures sont complémentaires

```
┌─────────────────────────────────────────────────────┐
│                   UCP Merchant Test                 │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Arena (Monolithe)          A2A (Microservices)    │
│  ├── Production            ├── POC / Démos         │
│  ├── Performance max       ├── Tests isolés        │
│  ├── BDD réelle            ├── Mock data           │
│  └── ADR 0001-0010         └── ADR 0011-0012       │
│                                                     │
│  Même logique métier (agents pricing)              │
│  Seule la couche transport diffère                 │
└─────────────────────────────────────────────────────┘
```

---

## 🔗 Ressources complémentaires

- **[Guide complet Agents A2A](../agents-a2a-guide.md)** - Tutorial, How-to, Reference, Explanation
- **[Navigation](../NAVIGATION.md)** - Trouvez rapidement ce que vous cherchez
- **[README principal](../../README.md)** - Point d'entrée du projet

---

## 📝 Template ADR

Pour créer un nouvel ADR, suivre la structure :

```markdown
# ADR-00XX: Titre de la décision

**Date**: YYYY-MM-DD
**Statut**: ✅ Accepté / 🚧 En discussion / ❌ Rejeté / 🔄 Supersédé par ADR-YYYY
**Décideurs**: Équipe Technique OCTO
**Tags**: `tag1`, `tag2`, `tag3`

## Contexte
Quel est le problème ? Quelle contrainte ?

## Décision
Quelle solution avons-nous choisie ?

## Conséquences
### Positives
- ✅ Avantage 1
- ✅ Avantage 2

### Négatives
- ❌ Inconvénient 1
- ⚠️ Mitigation : Comment on gère cet inconvénient

## Alternatives considérées
### 1. Alternative A
**Pour** : Avantages
**Contre** : Inconvénients
**Rejet** : Raison du rejet

## Métriques de succès
- ✅ Métrique 1 : Cible (atteint : valeur)

## Liens
- Code : `path/to/code`
- Docs : `path/to/docs`
- Commits : `hash1`, `hash2`
```

---

*Dernière mise à jour : 2024-06-10*
