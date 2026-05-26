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
| 002 | Protocole A2A pour Communication Inter-Agents | ✅ Accepté | 2026-03-11 | [Voir ADR →](002-a2a-protocol-inter-agent-communication.md) |

### Tableau Détaillé

| # | Titre | Décision Architecturale | Pourquoi c'est Important | Impact Concret |
|---|-------|------------------------|--------------------------|----------------|
| **[001](001-multi-agent-shopping-architecture.md)** | Architecture Multi-Agent Shopping | **Système distribué** avec 4 composants indépendants :<br><br>• Shopping Graph (recherche cross-merchant)<br>• Client Agent (Gemini, 8 tools)<br>• Observability Hub (dashboard SSE)<br>• 3 Merchants (SuperShop, MegaMart, BudgetBuy) | • Démontre l'utilité des protocoles A2A et MCP<br>• Permet la comparaison de prix cross-merchant<br>• Observabilité du raisonnement agent en temps réel<br>• Architecture extensible | • 4 binaires séparés à lancer<br>• Module `demo/` créé avec go.work<br>• Move `internal/` → `pkg/` pour réutilisabilité<br>• Évolutions : Arena mode, Ranking algorithm, Buying modes |
| **[002](002-a2a-protocol-inter-agent-communication.md)** | Protocole A2A pour Communication Inter-Agents | **A2A** comme standard de communication agent-to-agent<br><br>Alternatives rejetées :<br>• REST (pas de discovery standard)<br>• gRPC (complexité protobuf)<br>• Message Queue (over-engineering)<br>• MCP (sémantique LLM→Tool, pas Agent→Service) | • **Interopérabilité** : Spec ouverte (a2a.dev)<br>• **Sécurité** : OAuth2+PKCE (pas API keys)<br>• **Discovery** : Agent card JSON automatique<br>• **Session management** : Built-in pour contexte conversationnel | • Custom client A2A : ~400 lignes (auth.go, client.go, types.go)<br>• Server A2A : ~2000 lignes (executor, 16 handlers, tests)<br>• Shopping Graph polle catalogues via A2A<br>• Client Agent fait checkout/order via A2A |

### Lien entre les ADR

```
┌─────────────────────────────────────────┐
│  ADR 001 : Architecture Multi-Agent     │
│  → Définit QUOI (4 composants)          │
└──────────────────┬──────────────────────┘
                   │ nécessite
                   ▼
┌─────────────────────────────────────────┐
│  ADR 002 : Protocole A2A                │
│  → Définit COMMENT (communication)      │
└─────────────────────────────────────────┘
```

**ADR 001** = Décision système (distribué vs monolithique)  
**ADR 002** = Décision protocole (A2A vs REST/gRPC/MQ/MCP)

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
