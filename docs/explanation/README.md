# Explanation - Comprendre en Profondeur

Les articles d'explication explorent le **pourquoi** et le **comment** des décisions d'architecture. Ils fournissent le contexte, les trade-offs, et la vision globale du projet.

## Articles Disponibles

### Architecture & Design

#### [Pourquoi un Système Multi-Agents ?](why-multi-agent.md)
**Question** : Pourquoi décomposer le pricing en plusieurs agents au lieu d'un seul algorithme ?

**Ce que vous allez comprendre** :
- Les limites des systèmes monolithiques
- Les avantages de la séparation des préoccupations
- Pourquoi 3 agents (et pas 2 ou 10)
- Le pattern wrapper (Agent 3 et ses 4 sub-agents)
- Comparaison avec d'autres patterns (rule-based, ML, microservices)
- Comportements émergents (guerre des prix)

**À lire si** : Vous voulez comprendre la raison d'être de l'architecture.

#### [Trade-offs Compétitivité vs. Rentabilité](competitive-tradeoffs.md)
**Question** : Pourquoi le système accepte-t-il des marges négatives ?

**Ce que vous allez comprendre** :
- Le dilemme fondamental (gagner vs. marge)
- Les 4 stratégies (premium, balanced, aggressive, VIP retention)
- ADR-0002 : Victoire avant marge parfaite
- Analyse du cas réel (démo arena 5 juin, marge -8%)
- Solutions possibles (seuils, mode defensive, budget perte)
- Leçons apprises

**À lire si** : Vous vous demandez pourquoi les prix descendent si bas.

### Protocoles & Standards

#### [Intégration UCP/MCP/A2A](ucp-integration.md) *(À créer)*
**Question** : Pourquoi trois protocoles différents ?

**Ce que vous allez comprendre** :
- UCP pour standardisation e-commerce
- MCP pour tooling AI (Claude, etc.)
- A2A pour communication inter-agents
- Pourquoi multi-transport au lieu d'un seul
- Cas d'usage de chaque protocole

### System Design

#### [Philosophie d'Architecture](design-philosophy.md) *(À créer)*
**Question** : Quels principes guident le design du système ?

**Ce que vous allez comprendre** :
- Modularité > monolithe
- Explicabilité > performance
- Expérimentation > production-ready
- Event-driven > request-response
- Comportements émergents attendus

## Comment Lire ces Articles

Les articles d'explication sont **optionnels** pour utiliser le système, mais **essentiels** pour :
- Comprendre les décisions d'architecture
- Modifier le système intelligemment
- Débugger les comportements non-évidents
- Contribuer au projet

**Pas besoin de tout lire** : Picorez selon vos questions.

## Différence avec Tutorials et How-to

| Tutorial | How-to | Explanation |
|----------|--------|-------------|
| Apprendre à faire | Résoudre un problème | Comprendre pourquoi |
| Pas-à-pas pratique | Recette | Contexte & analyse |
| Pour débutants | Pour utilisateurs | Pour architectes |
| "Comment compiler" | "Comment configurer X" | "Pourquoi multi-agents" |

**Exemple** :
- **Tutorial** : "Lancer l'Arena" → vous fait manipuler
- **How-to** : "Configurer une stratégie" → vous dit quoi modifier
- **Explanation** : "Trade-offs compétitifs" → vous explique le dilemme sous-jacent

## Relation avec les ADRs

Les **ADRs** (Architecture Decision Records) dans [`docs/decisions/`](../decisions/) sont des **décisions** spécifiques.

Les **Explanations** sont des **analyses** plus larges qui contextualisent plusieurs ADRs.

**Exemple** :
- **ADR-0002** : "Nous acceptons marge < 10% pour gagner" (décision)
- **Explanation: Competitive Trade-offs** : "Voici pourquoi ce dilemme existe et comment on pourrait faire autrement" (analyse)

## Contribuer aux Explanations

Les articles d'explication doivent :
- **Être orientés compréhension** : Pas de commandes à exécuter
- **Fournir du contexte** : Historique, alternatives, trade-offs
- **Être analytiques** : "Pourquoi X et pas Y ?"
- **Relier les concepts** : Liens vers ADRs, tutorials, how-to

Structure recommandée :
1. La question centrale
2. Le problème ou dilemme
3. Différentes approches possibles
4. Notre choix et pourquoi
5. Conséquences (positives et négatives)
6. Situations où ce choix ne s'applique pas
7. Références et lectures complémentaires

**Ne mélangez pas** :
- ❌ Instructions étape-par-étape → How-to
- ❌ Découverte guidée → Tutorial
- ❌ Specs techniques → Reference
- ✅ Analyse, contexte, trade-offs → Explanation

## Pour Aller Plus Loin

Après avoir lu les explanations :
- Consultez les [ADRs](../decisions/) pour voir les décisions historiques
- Relisez les [Tutorials](../tutorials/) avec une meilleure compréhension
- Expérimentez en modifiant le code selon votre nouvelle compréhension

La vraie compréhension vient de la combinaison : **lire** (explanation) + **faire** (tutorial) + **analyser** (logs, comportements).
