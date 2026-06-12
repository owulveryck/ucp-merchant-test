# ADR-0011: Agents A2A Indépendants (Microservices)

**Date**: 2026-06-09  
**Statut**: ✅ Accepté  
**Décideurs**: Équipe Technique OCTO  
**Tags**: `architecture`, `a2a`, `microservices`, `json-rpc`

## Contexte

L'architecture existante (Arena) couple tous les agents dans une application monolithique :
- Déploiement global obligatoire (impossible de tester 1 seul agent)
- Dépendances multiples (Shopping Graph, Obs Hub, BDD)
- Setup complexe pour les démos clients (30+ minutes)
- Difficulté à scaler individuellement les agents

**Problème métier** : Les prospects abandonnent avant de voir une démo fonctionnelle.

## Décision

Créer des **agents autonomes** communiquant via **JSON-RPC 2.0** (Agent-to-Agent protocol).

### Implémentation

**Structure créée** :
```
pkg/a2a/              # Infrastructure A2A
  ├── types.go        # JSONRPCRequest, AgentIdentity, AgentResponse
  ├── agent.go        # Interface Agent commune
  └── server.go       # Serveur HTTP JSON-RPC

cmd/customer-growth-agent/
  ├── agent.go        # Wrapper A2A du CustomerGrowthAgent
  └── main.go         # Binaire standalone (port 9001)

cmd/competitiveness-agent/
  ├── agent.go        # Wrapper A2A du CompletivenessAgent
  ├── main.go         # Binaire standalone (port 9002)
  └── mock_price_source.go  # Données concurrents mock

cmd/agents-dashboard/
  ├── main.go         # Serveur web (port 8080)
  ├── handlers.go     # Proxy vers agents
  └── templates/index.html  # Interface interactive
```

**Protocole** :
- JSON-RPC 2.0 (standard IETF)
- Endpoints : `/a2a` (RPC), `/identity`, `/methods`, `/health`
- Transport : HTTP (évolutif vers WebSocket)

**Données** :
- Mock data intégré (4 clients, 4 produits avec concurrents)
- Pas de dépendance externe

## Conséquences

### Positives

**Business**
- ✅ POC en 30 secondes vs 30 minutes
- ✅ Démo sans risque technique (1 binaire = 1 agent)
- ✅ Déploiement progressif (1 agent → plateforme complète)

**Technique**
- ✅ Agents testables en isolation
- ✅ Scaling horizontal facile (N instances/agent)
- ✅ Crash isolation (1 agent down ≠ système down)
- ✅ Standard ouvert (pas de vendor lock-in)

**Coûts**
- ✅ Infrastructure réduite : 1 agent = 10 MB vs plateforme = GB
- ✅ Pas de BDD nécessaire pour démos

### Négatives

**Latence**
- ❌ Communication réseau entre agents (vs in-process)
- ⚠️ Mitigation : Pour production, option de compiler en monolithe reste possible

**Cohérence**
- ❌ Pas de transactions distribuées (chaque agent = état indépendant)
- ⚠️ Mitigation : Pour cas nécessitant ACID, utiliser orchestrateur externe

**Complexité opérationnelle**
- ❌ N binaires à déployer vs 1 monolithe
- ⚠️ Mitigation : Scripts de déploiement (`start-agents.sh`, `stop-agents.sh`)

## Métriques de succès

- ✅ Temps de setup démo : < 1 minute (atteint : 30s)
- ✅ Mémoire par agent : < 20 MB (atteint : ~10 MB)
- ✅ Taux de conversion prospects : +X% (à mesurer)

## Liens

- Code : `pkg/a2a/`, `cmd/*-agent/`
- Docs : `docs/agents-a2a-guide.md`, `docs/agents-a2a-summary.md`
- Spec JSON-RPC 2.0 : https://www.jsonrpc.org/specification
- Commits : `0883d41`, `a891f01`, `d6fa715`, `5d59900`

## Notes

Cette architecture A2A **complète** l'architecture Arena monolithique, elle ne la remplace pas :
- **A2A** : POCs, démos, tests isolés, environnements contraints
- **Arena** : Production haute performance, transactions complexes

Les deux approches partagent la même **logique métier** (agents), seule la couche transport diffère.
