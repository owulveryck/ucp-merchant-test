# Index des ADRs - UCP Merchant Test

## 📚 Par thématique

### 🤖 Multi-Agent & Orchestration

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0001](0001-architecture-multi-agents-pour-prix-competitif.md) | Architecture multi-agents pour prix compétitif | Décision d'utiliser plusieurs agents spécialisés | 2024-05-29 |
| [0004](0004-architecture-3-agents-orchestree.md) | Architecture 3 agents orchestrée | Price Intelligence, Market Analysis, Strategy Recommender | 2024-06-04 |
| [0005](0005-agent-acheteur-integre.md) | Agent acheteur intégré | Shopping Agent dans l'architecture Arena | 2024-06-04 |
| [0008](0008-multi-agent-shopping-architecture.md) | Multi-agent shopping architecture | Architecture complète Shopping Graph + Obs Hub | 2024-05-26 |
| [0010](0010-competitive-pricing-agent.md) | Competitive pricing agent | Agent de pricing compétitif 4-sous-agents | 2024-05-28 |
| [0011](0011-agents-a2a-independants.md) | Agents A2A Indépendants | **Microservices autonomes JSON-RPC 2.0** | 2024-06-09 |

### 💰 Stratégies de Pricing

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0002](0002-strategie-victoire-avant-marge-parfaite.md) | Stratégie victoire avant marge parfaite | Prioriser la compétitivité vs marge max | 2024-05-29 |
| [0003](0003-strategie-detection-codes-promo.md) | Stratégie détection codes promo | Recherche intelligente de codes promotionnels | 2024-05-29 |

### 📡 Transport & Communication

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0009](0009-multi-transport-architecture.md) | Multi-transport architecture | REST + MCP (JSON-RPC) + A2A | 2024-06-04 |
| [0011](0011-agents-a2a-independants.md) | Agents A2A Indépendants | **Protocole JSON-RPC 2.0 standard** | 2024-06-09 |

### 🎯 Observabilité & Messages

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0006](0006-messages-detailles-decision-achat.md) | Messages détaillés décision achat | Traçabilité des décisions agents | 2024-06-04 |

### 🎮 Scénarios & Tests

| ADR | Titre | Description | Date |
|-----|-------|-------------|------|
| [0007](0007-scenario-challenge-concurrents.md) | Scénario challenge concurrents | Environnement de test multi-agents | 2024-06-04 |
| [0012](0012-mock-data-sources-standalone.md) | Mock Data Sources standalone | **Données de test intégrées** | 2024-06-09 |

---

## 🔄 Par chronologie

```
2024-05-26
  └── 0008 - Multi-agent shopping architecture

2024-05-28
  ├── 0010 - Competitive pricing agent
  └── README_ADR.md

2024-05-29
  ├── 0001 - Architecture multi-agents pour prix compétitif
  ├── 0002 - Stratégie victoire avant marge parfaite
  └── 0003 - Stratégie détection codes promo

2024-06-04
  ├── 0004 - Architecture 3 agents orchestrée
  ├── 0005 - Agent acheteur intégré
  ├── 0006 - Messages détaillés décision achat
  ├── 0007 - Scénario challenge concurrents
  ├── 0009 - Multi-transport architecture
  └── README.md

2024-06-09 (Agents A2A 🆕)
  ├── 0011 - Agents A2A Indépendants
  └── 0012 - Mock Data Sources standalone
```

---

## 🎯 Par cas d'usage

### Je veux comprendre l'architecture globale
1. [ADR-0008: Multi-agent shopping architecture](0008-multi-agent-shopping-architecture.md)
2. [ADR-0009: Multi-transport architecture](0009-multi-transport-architecture.md)
3. [ADR-0011: Agents A2A Indépendants](0011-agents-a2a-independants.md)

### Je veux comprendre le pricing compétitif
1. [ADR-0001: Architecture multi-agents pour prix compétitif](0001-architecture-multi-agents-pour-prix-competitif.md)
2. [ADR-0004: Architecture 3 agents orchestrée](0004-architecture-3-agents-orchestree.md)
3. [ADR-0010: Competitive pricing agent](0010-competitive-pricing-agent.md)

### Je veux faire une démo rapide
1. [ADR-0011: Agents A2A Indépendants](0011-agents-a2a-independants.md) ⭐
2. [ADR-0012: Mock Data Sources standalone](0012-mock-data-sources-standalone.md)

### Je veux comprendre les stratégies business
1. [ADR-0002: Stratégie victoire avant marge parfaite](0002-strategie-victoire-avant-marge-parfaite.md)
2. [ADR-0003: Stratégie détection codes promo](0003-strategie-detection-codes-promo.md)

---

## 📊 Statistiques

- **Total ADRs** : 12
- **Architecture** : 6 ADRs (0001, 0004, 0005, 0008, 0009, 0011)
- **Stratégies** : 2 ADRs (0002, 0003)
- **Tests** : 2 ADRs (0007, 0012)
- **Observabilité** : 1 ADR (0006)
- **Agents** : 2 ADRs (0010, 0011)

---

## 🆕 Nouveautés (Juin 2024)

### ADR-0011: Agents A2A Indépendants
**Changement majeur** : Passage de l'architecture monolithique Arena à des microservices autonomes.

**Avant** (Arena) :
- Tous les agents couplés dans 1 application
- Déploiement global obligatoire
- Setup complexe (30+ minutes)

**Après** (A2A) :
- Chaque agent = microservice indépendant
- Déploiement sélectif (1 agent à la fois)
- Setup instantané (30 secondes)

**Impact** :
- ✅ POC en jours vs semaines
- ✅ Démos clients sans risque technique
- ✅ Coûts infrastructure réduits

**Lien** : [ADR-0011 complet](0011-agents-a2a-independants.md)

### ADR-0012: Mock Data Sources
**Complément** : Données de test intégrées pour agents standalone.

**Données fournies** :
- 4 clients de test (elsi, alice, bob, john)
- 4 produits avec concurrents (laptop, mouse, keyboard, monitor)
- Zéro configuration nécessaire

**Impact** :
- ✅ Tests reproductibles
- ✅ Démos offline (train, avion)
- ✅ Onboarding développeurs instantané

**Lien** : [ADR-0012 complet](0012-mock-data-sources-standalone.md)

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
