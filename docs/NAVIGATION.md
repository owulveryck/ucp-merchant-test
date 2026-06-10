# 🧭 Navigation du Repository

## 🎯 Par objectif

### Je veux tester rapidement les agents A2A
1. **[Tutorial 5 minutes](agents-a2a-guide.md#-tutorial---premier-lancement-apprentissage)** ← COMMENCEZ ICI
2. Lancez : `./scripts/start-agents.sh`
3. Ouvrez : http://localhost:8080

### Je veux comprendre l'architecture
1. **[ADR-0011: Agents A2A](decisions/0011-agents-a2a-independants.md)** - Pourquoi des microservices ?
2. **[Concepts expliqués](agents-a2a-guide.md#-explanation---comprendre-les-concepts)** - Comment ça marche ?
3. **[Diagramme architecture](agents-a2a-summary.md#️-architecture)** - Vue d'ensemble

### Je veux intégrer les agents dans mon projet
1. **[Reference API](agents-a2a-guide.md#-reference---documentation-technique)** - Tous les endpoints
2. **[How-to: Découvrir méthodes](agents-a2a-guide.md#comment-découvrir-les-méthodes-dun-agent)**
3. **Code source** : `pkg/a2a/` (infrastructure) + `cmd/*-agent/` (agents)

### Je veux démontrer aux clients
1. **[Valeur business](agents-a2a-summary.md#-intérêt-pour-les-clients)** - Arguments commerciaux
2. **[Exemple concret](agents-a2a-summary.md#-exemple-concret)** - 30s vs 30min
3. **[Dashboard démo](http://localhost:8080)** - Interface visuelle

---

## 📂 Par type de contenu

### 🎓 Apprentissage (Tutorial)
**Objectif** : Apprendre en faisant, étape par étape

| Document | Description | Temps |
|----------|-------------|-------|
| [Tutorial Agents A2A](agents-a2a-guide.md#-tutorial---premier-lancement-apprentissage) | Votre premier agent en 5 minutes | ⏱️ 5 min |
| [Exemple concret](agents-a2a-summary.md#-exemple-concret) | Scénario client réel | ⏱️ 2 min |

### 🔧 Pratique (How-to)
**Objectif** : Résoudre un problème spécifique

| Document | Description |
|----------|-------------|
| [How-to Guides](agents-a2a-guide.md#-how-to-guides---tâches-pratiques) | Toutes les tâches courantes |
| ↳ Lancer tous les agents | `./scripts/start-agents.sh` |
| ↳ Tester l'agent de compétitivité | Requête curl exemple |
| ↳ Ajouter un client de test | Modifier `mock_customer_data.go` |
| ↳ Changer le port d'un agent | Flag `--port` |

### 📖 Référence (Reference)
**Objectif** : Trouver une info technique précise

| Document | Description |
|----------|-------------|
| [Reference API](agents-a2a-guide.md#-reference---documentation-technique) | Tous les endpoints détaillés |
| ↳ Customer Growth Agent | Méthodes, paramètres, exemples |
| ↳ Competitiveness Agent | Méthodes, paramètres, exemples |
| ↳ Dashboard Web | API et interface |
| [Clients de test](agents-a2a-guide.md#clients-de-test-disponibles) | elsi, alice, bob, john |
| [Produits mock](agents-a2a-guide.md#produits-avec-données-concurrents) | laptop, mouse, keyboard, monitor |
| [JSON-RPC 2.0](agents-a2a-guide.md#structure-json-rpc-20) | Format requête/réponse |

### 💡 Explication (Explanation)
**Objectif** : Comprendre les concepts et décisions

| Document | Description |
|----------|-------------|
| [Concepts A2A](agents-a2a-guide.md#-explanation---comprendre-les-concepts) | Pourquoi ? Comment ? |
| ↳ Monolithe vs Microservices | Comparaison architecturale |
| ↳ Communication inter-agents | Protocole JSON-RPC |
| ↳ Mock Data vs Production | Stratégie de données |
| [ADR-0011: Agents A2A](decisions/0011-agents-a2a-independants.md) | Décision architecture microservices |
| [ADR-0012: Mock Data](decisions/0012-mock-data-sources-standalone.md) | Décision données intégrées |

---

## 👤 Par profil utilisateur

### Développeur
**Ce qui vous intéresse** : Code, API, tests

1. **Démarrage rapide** : [Tutorial 5 min](agents-a2a-guide.md#-tutorial---premier-lancement-apprentissage)
2. **Code source** :
   - Infrastructure : `pkg/a2a/`
   - Agents : `cmd/customer-growth-agent/`, `cmd/competitiveness-agent/`
   - Dashboard : `cmd/agents-dashboard/`
3. **API Reference** : [Endpoints et paramètres](agents-a2a-guide.md#-reference---documentation-technique)
4. **Ajouter données** : [How-to ajouter client](agents-a2a-guide.md#comment-ajouter-un-nouveau-client-de-test)

### Commercial / Sales
**Ce qui vous intéresse** : Valeur client, démo rapide

1. **Arguments business** : [5 raisons d'adopter A2A](agents-a2a-summary.md#-intérêt-pour-les-clients)
2. **Démo 30 secondes** : [Exemple concret](agents-a2a-summary.md#-exemple-concret)
3. **Différence vs concurrent** : [Comparaison architecture](agents-a2a-summary.md#-différence-avec-le-système-monolithique)
4. **Interface visuelle** : [Dashboard](http://localhost:8080) (après `./scripts/start-agents.sh`)

### Architecte
**Ce qui vous intéresse** : Décisions, patterns, scalabilité

1. **ADRs complets** :
   - [ADR-0011: Microservices A2A](decisions/0011-agents-a2a-independants.md)
   - [ADR-0012: Mock Data Sources](decisions/0012-mock-data-sources-standalone.md)
2. **Architecture détaillée** : [Concepts expliqués](agents-a2a-guide.md#-explanation---comprendre-les-concepts)
3. **Alternatives rejetées** : Voir sections "Alternatives" dans les ADRs
4. **Migration production** : [Mock → Production](decisions/0012-mock-data-sources-standalone.md#migration-vers-production)

### Chef de projet
**Ce qui vous intéresse** : Planning, ROI, risques

1. **Valeur business** : [Intérêt clients](agents-a2a-summary.md#-intérêt-pour-les-clients)
   - ✅ Time-to-Market accéléré
   - ✅ Coûts infra réduits
   - ✅ POC en jours vs semaines
2. **Métriques** : Voir "Métriques de succès" dans [ADR-0011](decisions/0011-agents-a2a-independants.md#métriques-de-succès)
3. **Évolution future** : [Roadmap](agents-a2a-guide.md#évolution-future)
4. **Risques** : Voir "Conséquences négatives" dans les ADRs

---

## 🗂️ Structure du repository

```
ucp-merchant-test/
├── cmd/                                    # Applications (binaires)
│   ├── customer-growth-agent/              # Agent Fidélisation (port 9001)
│   ├── competitiveness-agent/              # Agent Stratégie Prix (port 9002)
│   └── agents-dashboard/                   # Dashboard Web (port 8080)
│
├── pkg/                                    # Bibliothèques réutilisables
│   ├── a2a/                                # Infrastructure Agent-to-Agent
│   │   ├── types.go                        # JSON-RPC structures
│   │   ├── agent.go                        # Interface Agent
│   │   └── server.go                       # Serveur HTTP
│   └── pricing-unified/                    # Système pricing
│       ├── agents/                         # Logique métier agents
│       └── datasources/                    # Sources de données
│           └── mock_customer_data.go       # Clients de test
│
├── docs/                                   # Documentation
│   ├── NAVIGATION.md                       # 👈 VOUS ÊTES ICI
│   ├── agents-a2a-guide.md                 # Guide complet (Divio)
│   ├── agents-a2a-summary.md               # Résumé exécutif
│   └── decisions/                          # Architecture Decision Records
│       ├── 0011-agents-a2a-independants.md
│       └── 0012-mock-data-sources-standalone.md
│
├── scripts/                                # Scripts utilitaires
│   ├── start-agents.sh                     # Lance tous les agents
│   └── stop-agents.sh                      # Arrête tous les agents
│
└── bin/                                    # Binaires compilés (gitignore)
    ├── customer-growth-agent
    ├── competitiveness-agent
    └── agents-dashboard
```

---

## 🔗 Liens externes utiles

- **Spec JSON-RPC 2.0** : https://www.jsonrpc.org/specification
- **Universal Commerce Protocol** : https://ucp.dev
- **Framework Divio** : https://documentation.divio.com
- **GitHub du projet** : https://github.com/owulveryck/ucp-merchant-test/tree/stageocto

---

## ❓ Questions fréquentes

**Q: Par où commencer si je n'ai jamais utilisé les agents A2A ?**  
R: [Tutorial 5 minutes](agents-a2a-guide.md#-tutorial---premier-lancement-apprentissage) ← Commencez ici !

**Q: Quelle est la différence entre les agents A2A et le système Arena ?**  
R: [Comparaison détaillée](agents-a2a-summary.md#-différence-avec-le-système-monolithique)

**Q: Comment démontrer aux clients en moins de 2 minutes ?**  
R: [Exemple concret](agents-a2a-summary.md#-exemple-concret) - Script de démo prêt à l'emploi

**Q: Les données de test sont-elles suffisantes pour une démo ?**  
R: Oui ! 4 clients (elsi, alice, bob, john) + 4 produits. [Voir détails](agents-a2a-guide.md#clients-de-test-disponibles)

**Q: Comment ajouter mes propres clients de test ?**  
R: [How-to ajouter client](agents-a2a-guide.md#comment-ajouter-un-nouveau-client-de-test)

**Q: Peut-on déployer en production ?**  
R: Oui, voir [Migration production](decisions/0012-mock-data-sources-standalone.md#migration-vers-production)

**Q: Quelle est la roadmap future ?**  
R: [Évolution future](agents-a2a-guide.md#évolution-future) - Service Discovery, Load Balancing, Auth...

---

## 📝 Contribuer

Pour ajouter de nouvelles fonctionnalités ou agents :

1. **Créer un ADR** : `docs/decisions/00XX-votre-decision.md`
2. **Suivre le pattern** : Voir structure dans `cmd/customer-growth-agent/`
3. **Documenter** : Ajouter section dans `agents-a2a-guide.md`
4. **Tester** : Démarrer l'agent et vérifier les endpoints

**Template ADR** : Voir [ADR-0011](decisions/0011-agents-a2a-independants.md) comme exemple.
