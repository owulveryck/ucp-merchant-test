# Architecture Decision Records (ADR)

Ce répertoire contient les **Architecture Decision Records** du projet UCP merchant test.

## Qu'est-ce qu'un ADR ?

Un ADR documente une **décision architecturale significative** : un choix de design qui :
- Adresse une exigence fonctionnelle ou non-fonctionnelle importante
- A un impact structurel sur le système
- Est difficile à changer une fois implémentée
- Implique des trade-offs importants

Les ADR capturent le **contexte**, les **alternatives considérées**, et les **conséquences** de chaque décision.

---

## ADR du Projet

### Vue d'ensemble

| # | Titre | Statut | Date | Lien |
|---|-------|--------|------|------|
| 001 | Architecture Multi-Agent Shopping | ✅ Accepté | 2026-03-11 | [Voir ADR →](001-multi-agent-shopping-architecture.md) |
| 002 | Architecture Multi-Transport (REST, MCP, A2A) | ✅ Accepté | 2026-03-10-11 | [Voir ADR →](002-multi-transport-architecture.md) |

### Tableau Détaillé

| # | Titre | Décision Architecturale | Pourquoi c'est Important | Impact Concret |
|---|-------|------------------------|--------------------------|----------------|
| **[001](001-multi-agent-shopping-architecture.md)** | Architecture Multi-Agent Shopping | **Système distribué** avec 4 composants indépendants :<br><br>• Shopping Graph (recherche cross-merchant)<br>• Client Agent (Gemini, 8 tools)<br>• Observability Hub (dashboard SSE)<br>• 3 Merchants (SuperShop, MegaMart, BudgetBuy) | • Démontre l'utilité des protocoles A2A et MCP<br>• Permet la comparaison de prix cross-merchant<br>• Observabilité du raisonnement agent en temps réel<br>• Architecture extensible | • 4 binaires séparés à lancer<br>• Module `demo/` créé avec go.work<br>• Move `internal/` → `pkg/` pour réutilisabilité<br>• Évolutions : Arena mode, Ranking algorithm, Buying modes |
| **[002](002-multi-transport-architecture.md)** | Architecture Multi-Transport (REST, MCP, A2A) | **3 protocoles simultanés** pour différents clients :<br><br>• **REST** : Web/mobile, tests, debug<br>• **MCP** : Claude Desktop, IDEs, LLM clients<br>• **A2A** : Shopping Graph, Client Agent (agents autonomes) | • **Flexibilité** : Chaque client choisit son protocole<br>• **Zero duplication** : Tous délèguent à `merchant.Merchant`<br>• **Extensibilité** : Pattern établi pour ajouter GraphQL/gRPC<br>• **Conformité** : UCP (REST), MCP, A2A specs respectées | • REST : ~400 LOC (endpoints HTTP)<br>• MCP : ~900 LOC (via mcp-go library)<br>• A2A : ~2400 LOC (client 400 + server 2000)<br>• Tests : 60 UCP + 43 MCP + unit tests A2A<br>• Architecture validée par ajout A2A sans refonte |

### Lien entre les ADR

```
┌─────────────────────────────────────────┐
│  ADR 001 : Architecture Multi-Agent     │
│  → Définit QUOI (4 composants)          │
└──────────────────┬──────────────────────┘
                   │ nécessite
                   ▼
┌─────────────────────────────────────────┐
│  ADR 002 : Multi-Transport              │
│  → Définit COMMENT (3 protocoles)       │
│  REST (web) + MCP (LLM) + A2A (agents)  │
└─────────────────────────────────────────┘
```

**ADR 001** = Décision système (distribué vs monolithique)  
**ADR 002** = Décision transports (3 protocoles pour 3 types de clients)

---

## Statuts Possibles

- **Proposé** : En discussion, non implémenté
- **✅ Accepté** : Implémenté et actif
- **⚠️ Déprécié** : Remplacé par une meilleure approche mais encore présent dans le code
- **❌ Rejeté** : Alternative considérée mais non retenue
- **🔄 Remplacé** : Remplacé par un autre ADR (référence dans le document)

---

## Quand Créer un Nouvel ADR ?

Créez un ADR quand vous prenez une décision qui répond à **au moins 3** de ces critères :

✅ **Structurelle** : Affecte l'organisation globale du système (architecture, modules, composants)  
✅ **Difficile à changer** : Refactorer cette décision nécessiterait un effort significatif  
✅ **Trade-offs importants** : Il y a des avantages ET inconvénients à documenter  
✅ **Multi-composants** : Affecte plusieurs parties du système  
✅ **Quality attributes** : Impact sur performance, scalabilité, sécurité, maintenabilité  

### ❌ Ne créez PAS d'ADR pour :

- Choix de librairies mineures (logger, formatter)
- Patterns de design locaux (Factory, Strategy dans un seul package)
- Features produit (sauf si elles nécessitent des choix architecturaux)
- Conventions de code (mettre ça dans un CONTRIBUTING.md)

---

## Template ADR

```markdown
# ADR XXX : [Titre de la Décision]

- **Date** : YYYY-MM-DD
- **Statut** : Proposé | Accepté | Déprécié | Rejeté
- **Décideurs** : [Noms]
- **Lié à** : [ADR-YYY] (optionnel)

## Contexte

Décrire :
- Le problème ou l'exigence (fonctionnelle ou non-fonctionnelle)
- Pourquoi cette décision doit être prise maintenant
- Les contraintes (techniques, temps, budget, compétences)

## Décision

Quelle solution a été choisie et pourquoi ? Décrire de manière claire et concise.

## Alternatives Considérées

### Alternative 1 : [Nom]

**Pour** :
- ✅ Avantage 1
- ✅ Avantage 2

**Contre** :
- ❌ Inconvénient 1
- ❌ Inconvénient 2

**Verdict** : Rejeté/Accepté. Justification.

## Trade-offs

### Positifs
- ✅ Bénéfice 1
- ✅ Bénéfice 2

### Négatifs
- ❌ Inconvénient 1
- ❌ Inconvénient 2

### Risques et Mitigations
- **Risque** : Description → Mitigation : Solution

## Conséquences

Impact concret sur :
- L'architecture
- Les composants existants
- Les développements futurs

## Évolutions Post-Implémentation (optionnel)

À remplir après quelques semaines/mois :
- Ce qui fonctionne bien
- Ce qui ne fonctionne pas comme prévu
- Leçons apprises
```

---

## Ressources

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions (Michael Nygard)](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR Tools](https://github.com/npryce/adr-tools)

---

## Contribuer

Pour proposer un nouvel ADR :

1. Copiez le template ci-dessus
2. Créez `docs/adr/XXX-titre-kebab-case.md` (numéro séquentiel suivant)
3. Remplissez toutes les sections (surtout Contexte, Alternatives, Trade-offs)
4. Soumettez en PR avec statut **Proposé**
5. Après discussion et implémentation, changez le statut en **Accepté**

**Rappel** : Un bon ADR documente le **pourquoi**, pas juste le **quoi**. Le code montre le "quoi", l'ADR explique la décision.
