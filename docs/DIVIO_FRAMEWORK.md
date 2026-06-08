# Documentation Divio Framework

Ce document explique comment la documentation du projet UCP Merchant Test est organisée selon le [framework Divio](https://documentation.divio.com/).

## Vue d'Ensemble

```
docs/
├── README.md                    # Point d'entrée (vous êtes ici)
├── DIVIO_FRAMEWORK.md          # Ce document
│
├── tutorials/                   # 📚 LEARNING - Apprendre en pratiquant
│   ├── README.md
│   ├── 01-getting-started.md
│   ├── 02-first-demo.md
│   └── 03-multi-agent-pricing.md
│
├── how-to/                      # 🔧 TASK - Résoudre des problèmes
│   ├── README.md
│   ├── run-conformance-tests.md
│   └── configure-merchant.md
│
├── reference/                   # 📖 INFORMATION - Consulter
│   ├── README.md
│   └── api-reference.md
│
├── explanation/                 # 💡 UNDERSTANDING - Comprendre
│   ├── README.md
│   ├── why-multi-agent.md
│   └── competitive-tradeoffs.md
│
└── decisions/                   # 📋 ADRs (hors Divio)
    ├── README.md
    └── 0001-*.md ... 0010-*.md
```

## Les 4 Types de Documentation

### 📚 Tutorials (Learning-Oriented)

**Pour qui** : Débutants découvrant le projet

**Objectif** : Apprendre en faisant

**Caractéristiques** :
- Pas-à-pas guidé
- Résultat concret à la fin
- Explications pédagogiques
- Pas d'options, un seul chemin

**Analogie** : Cours de cuisine

**Exemple** : "Tutorial : Getting Started" → vous fait compiler et lancer le serveur

**Ce qu'on N'Y met PAS** :
- Explications théoriques longues → Explanation
- Solutions à des problèmes spécifiques → How-to
- Specs techniques → Reference

### 🔧 How-to Guides (Task-Oriented)

**Pour qui** : Utilisateurs confirmés avec un problème précis

**Objectif** : Accomplir une tâche spécifique

**Caractéristiques** :
- Orienté résultat
- Instructions concises
- Suppose connaissance de base
- Solutions pratiques

**Analogie** : Recette de cuisine

**Exemple** : "How-to : Run Conformance Tests" → étapes pour lancer les tests

**Ce qu'on N'Y met PAS** :
- Parcours complet d'apprentissage → Tutorial
- Explication du pourquoi → Explanation
- Liste exhaustive des options → Reference

### 📖 Reference (Information-Oriented)

**Pour qui** : Tous, pour consultation rapide

**Objectif** : Trouver une information précise

**Caractéristiques** :
- Format encyclopédique
- Exhaustif et précis
- Exemples concrets
- Structure prévisible

**Analogie** : Dictionnaire / Manuel technique

**Exemple** : "API Reference" → tous les endpoints avec paramètres et exemples

**Ce qu'on N'Y met PAS** :
- Parcours narratif → Tutorial
- Recettes → How-to
- Analyses et contexte → Explanation

### 💡 Explanation (Understanding-Oriented)

**Pour qui** : Architectes, contributeurs, curieux

**Objectif** : Comprendre le contexte et les décisions

**Caractéristiques** :
- Analyse en profondeur
- Trade-offs et alternatives
- Contexte historique
- Vision globale

**Analogie** : Article scientifique / Essai

**Exemple** : "Why Multi-Agent" → raisons architecturales, comparaisons, trade-offs

**Ce qu'on N'Y met PAS** :
- Instructions pratiques → How-to / Tutorial
- Specs techniques brutes → Reference

## Tableau de Décision

| Je veux... | Type | Exemple |
|-----------|------|---------|
| Apprendre le projet de zéro | Tutorial | "Getting Started" |
| Lancer la démo shopping | Tutorial | "First Demo" |
| Exécuter les tests UCP | How-to | "Run Conformance Tests" |
| Ajouter un marchand | How-to | "Configure Merchant" |
| Connaître les endpoints API | Reference | "API Reference" |
| Comprendre un champ de données | Reference | "Data Models" |
| Savoir pourquoi multi-agents | Explanation | "Why Multi-Agent" |
| Comprendre marge négative | Explanation | "Competitive Trade-offs" |

## Principes Clés

### 1. Séparation Stricte

Chaque document appartient à **un seul** type. Ne mélangez pas :
- ❌ Tutorial avec référence technique exhaustive
- ❌ How-to avec explication longue du contexte
- ❌ Reference avec parcours d'apprentissage
- ❌ Explanation avec instructions pas-à-pas

### 2. Liens Croisés

Les documents doivent se référencer :
- Tutorial → renvoie vers How-to pour approfondir
- How-to → renvoie vers Reference pour détails
- Explanation → renvoie vers Tutorial pour pratiquer
- Reference → renvoie vers Explanation pour contexte

### 3. Autonomie

Chaque document doit être **utilisable seul** :
- Indiquer les prérequis clairement
- Liens vers ressources externes si besoin
- Pas de dépendances implicites

### 4. Maintien

La documentation doit évoluer avec le code :
- Nouveau feature → Update Reference + Create How-to
- Changement architecture → Update Explanation + ADR
- Breaking change → Update Tutorials

## Workflow de Contribution

### Ajouter un Tutorial

1. Identifiez un parcours d'apprentissage manquant
2. Créez `tutorials/0X-titre.md`
3. Structure : Objectif → Étapes → Résultat → Prochaines étapes
4. Testez le tutorial de zéro (vraiment!)
5. Ajoutez au `tutorials/README.md`

### Ajouter un How-to

1. Identifiez une tâche fréquente ou complexe
2. Créez `how-to/nom-de-la-tache.md`
3. Structure : Contexte → Prérequis → Étapes → Vérification → Troubleshooting
4. Testez la procédure
5. Ajoutez au `how-to/README.md`

### Ajouter une Reference

1. Identifiez info technique manquante ou dispersée
2. Créez `reference/nom-du-sujet.md`
3. Structure : Vue d'ensemble → Index → Sections détaillées → Exemples
4. Vérifiez exhaustivité dans le code source
5. Ajoutez au `reference/README.md`

### Ajouter une Explanation

1. Identifiez une décision/concept non-évident
2. Créez `explanation/nom-du-concept.md`
3. Structure : Question → Problème → Approches → Notre choix → Conséquences
4. Reliez aux ADRs si applicable
5. Ajoutez au `explanation/README.md`

## Exemples Concrets

### Mauvais : Tutorial qui devient Reference

```markdown
# Tutorial : Getting Started

## API Endpoints

### POST /checkout-sessions

Paramètres :
- items (array, required) - Liste des items
  - item_id (string, required) - ID du produit
  - quantity (int, required) - Quantité (> 0)
  - ...
(20 pages de specs)
```

**Problème** : Trop technique, perd l'utilisateur débutant.

**Solution** : Tutorial montre UN exemple simple, Reference liste tout.

### Bon : Séparation claire

**Tutorial** :
```markdown
Créez votre première checkout :
\`\`\`bash
curl -X POST ... -d '{"items": [{"item_id": "rose_bouquet", "quantity": 2}]}'
\`\`\`

Vous devriez voir un JSON avec id et status "incomplete".
Pour plus de détails sur l'API, voir [API Reference](../reference/api-reference.md).
```

**Reference** :
```markdown
### POST /checkout-sessions
...
(specs complètes)
```

## Métriques de Qualité

Une bonne documentation Divio devrait :
- ✅ Un débutant peut démarrer avec juste les tutorials
- ✅ Un utilisateur confirmé trouve rapidement dans how-to
- ✅ La reference répond à toute question technique
- ✅ Les explanations donnent le contexte des décisions

Si un utilisateur :
- Dit "Je ne sais pas par où commencer" → Manque tutorial
- Dit "Comment je fais X ?" → Manque how-to
- Dit "Quel est le format de Y ?" → Manque reference
- Dit "Pourquoi c'est fait comme ça ?" → Manque explanation

## Ressources

- [Divio Documentation System](https://documentation.divio.com/)
- [Exemple : Django Docs](https://docs.djangoproject.com/) (suit Divio)
- [Article : Grand Unified Theory of Documentation](https://www.writethedocs.org/videos/eu/2017/the-four-kinds-of-documentation-and-why-you-need-to-understand-what-they-are-daniele-procida/)

## Statut Actuel

✅ **Créé** :
- Structure de base (4 dossiers + READMEs)
- 3 tutorials complets
- 2 how-to guides
- 1 API reference
- 2 explanations

🚧 **À créer** :
- Reference : Configuration, Agent Architecture, Data Models, Error Codes
- How-to : Setup Pricing Strategies, Monitor Agent Decisions
- Explanation : UCP Integration, Design Philosophy

📝 **En continu** :
- Maintenir à jour avec le code
- Ajouter selon besoins utilisateurs
- Améliorer clarté basé sur feedback
