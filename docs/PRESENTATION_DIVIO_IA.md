# Documentation Structurée avec IA et Framework Divio

**Présentation pour consultants tech**  
**Durée : 10-15 minutes**

---

## SLIDE 1 : Le Problème

### Documentation = Chaos

**Symptômes classiques** :
- ❌ Tout mélangé dans un gros README
- ❌ "Comment installer ?" à côté de "Architecture interne"
- ❌ Tutorials mélangés avec référence technique
- ❌ Utilisateur perdu : "Par où commencer ?"

**Coût réel** :
- Questions répétées aux développeurs
- Onboarding lent (plusieurs jours au lieu d'heures)
- Adoption faible du projet
- Frustration équipe + utilisateurs

**Question** : Comment structurer la documentation de manière professionnelle, rapidement ?

---

## SLIDE 2 : Framework Divio - Les 4 Types

### Divio : Système de Documentation en 4 Quadrants

| Type | Orientation | Analogie | Exemple |
|------|-------------|----------|---------|
| 📚 **TUTORIALS** | Learning | Cours de cuisine | "Démarrage en 15min" |
| 🔧 **HOW-TO** | Task | Recette | "Configurer X" |
| 📖 **REFERENCE** | Information | Encyclopédie | "API endpoints" |
| 💡 **EXPLANATION** | Understanding | Article scientifique | "Pourquoi multi-agents" |

### Règle d'Or

**1 document = 1 type** (ne jamais mélanger)

### Pourquoi ça marche ?

- Utilisateur sait **où chercher** selon son besoin
- Rédacteur sait **quoi écrire** et **comment**
- Navigation **prédictible** et **intuitive**

**Ressource** : https://documentation.divio.com/

---

## SLIDE 3 : Comment l'IA Aide

### IA = Accélérateur Divio

**4 Super-Pouvoirs de l'IA** :

✅ **COMPREND** le framework mieux qu'un humain
- Applique les règles strictement
- Sépare les types automatiquement
- Identifie ce qui va où

✅ **GÉNÈRE** structure + contenu initial
- De zéro à arborescence complète en secondes
- Créé fichiers, dossiers, READMEs
- Contenu cohérent dès le départ

✅ **MAINTIENT** la cohérence
- Même ton partout
- Même format
- Même niveau de détail

✅ **ITÈRE** rapidement
- "Supprime section X" → fait en 10s
- "Ajoute exemple Y" → fait en 20s
- Zéro friction pour raffiner

### Principe Clé

**L'humain guide** (framework Divio + connaissance métier)  
**L'IA exécute** (rapidité + cohérence)

---

## SLIDE 4 : Démo Live - Prompt Initial

### Conversation avec l'IA (Exemple Réel)

```
UTILISATEUR :
"Je voudrais documenter le projet avec le framework Divio"

IA : [Analyse le projet en lisant README + structure]

→ Crée docs/
    ├── tutorials/
    ├── how-to/
    ├── reference/
    └── explanation/

→ Génère README.md avec navigation claire
→ Propose structure détaillée

UTILISATEUR :
"Ok, génère les contenus"

IA : [30-40 secondes par fichier]

✅ tutorials/01-getting-started.md
   → Pas-à-pas pour débutant (15min)

✅ how-to/run-tests.md
   → Recette précise pour exécuter tests

✅ reference/api-reference.md
   → Specs complètes de l'API

✅ explanation/why-multi-agent.md
   → Contexte et raisons architecturales
```

**Résultat** : Structure professionnelle en quelques prompts

---

## SLIDE 5 : Démo Live - Raffinement

### Itération en Temps Réel

```
UTILISATEUR :
"Supprime les sections 'à venir' dans tous les READMEs"

IA : [10 secondes]
✅ Édite 3 fichiers (explanation/, how-to/, reference/)
✅ Crée commit git avec message descriptif
✅ Push sur GitHub

────────────────────────────────────────

UTILISATEUR :
"Ajoute un exemple concret dans le tutorial 1"

IA : [20 secondes]
✅ Ajoute section avec code fonctionnel
✅ Vérifie cohérence avec le reste
```

### Avantages

- Zéro friction
- Zéro erreur manuelle
- Modifications atomiques
- Historique git propre

---

## SLIDE 6 : Cas Réel - Projet UCP

### Transformation Concrète

**AVANT** (état initial)
```
├── README.md (fourre-tout, 200 lignes)
└── docs/
    └── decisions/ (10 ADRs)
```

**APRÈS** (45-60 minutes avec IA)
```
docs/
├── README.md (navigation claire)
├── DIVIO_FRAMEWORK.md (guide)
├── PRESENTATION.md (slides)
├── tutorials/
│   ├── 01-getting-started.md
│   ├── 02-first-demo.md
│   └── 03-multi-agent-pricing.md
├── how-to/
│   ├── run-conformance-tests.md
│   └── configure-merchant.md
├── reference/
│   └── api-reference.md
├── explanation/
│   ├── why-multi-agent.md
│   └── competitive-tradeoffs.md
└── decisions/ (10 ADRs existants)
```

**Résultat** : 15 fichiers, ~2,500 lignes, structure professionnelle

---

## SLIDE 7 : Exemple Concret - Séparation des Types

### Même Sujet, 2 Approches Différentes

**❌ MAUVAIS** (tout mélangé dans README)
```markdown
# API Checkout

Pour créer un checkout, faites POST /checkout-sessions.
Voici comment ça marche : le checkout permet de...
Exemple : curl -X POST http://localhost:8182/...

Paramètres :
- items (array, required) - liste des items
  - item_id (string, required) - ID du produit
  - quantity (int, required) - quantité
  ... [20 pages de specs]

Pourquoi on a choisi ce design : blabla architecture...
```

**✅ BON** (séparé selon Divio)

**Tutorial** (`tutorials/01-getting-started.md`)
```markdown
## Créer votre première checkout

Testez l'API avec curl :
\`\`\`bash
curl -X POST http://localhost:8182/checkout-sessions \\
  -d '{"items": [{"item_id": "rose_bouquet", "quantity": 2}]}'
\`\`\`

Vous devriez voir un JSON avec status "incomplete".
```

**Reference** (`reference/api-reference.md`)
```markdown
### POST /checkout-sessions

Paramètres :
- items (array, required) - liste des items
  - item_id (string, required) - ID produit
  - quantity (int, required) - quantité (> 0)

Response 201 : {...}
```

**Explanation** (`explanation/why-checkout-before-order.md`)
```markdown
# Pourquoi Checkout Avant Order ?

Le pattern checkout → order permet...
Trade-offs : ...
```

→ **Séparation claire = utilisateur trouve immédiatement**

---

## SLIDE 8 : Métriques - ROI Mesuré

### Gain de Temps (Session Réelle)

**📊 Production IA** (notre session)
- ✅ **15 fichiers** créés
- ✅ **~2,500 lignes** de documentation
- ✅ **2 commits** git propres
- ⏱️ **45-60 minutes** de session

**📊 Équivalent Manuel Estimé**
- 📝 Rédaction : **2 jours** (écrire 2,500 lignes)
- 🏗️ Structure : **0.5 jour** (concevoir arborescence)
- ✅ Cohérence : **0.5 jour** (relecture, harmonisation)
- **Total : ~3 jours**

### ROI

**→ GAIN : 95% de temps**  
**→ QUALITÉ : Cohérence parfaite**  
**→ MAINTENANCE : Structure claire pour évolutions futures**

---

## SLIDE 9 : Pièges à Éviter

### ⚠️ Ce Qui Peut Mal Tourner

**1. IA mélange parfois les types**
```
❌ Tutorial qui devient référence technique exhaustive
✅ Solution : "C'est un tutorial, pas de specs complètes ici"
```

**2. Exemples peuvent être génériques**
```
❌ Code exemple fictif qui ne compile pas
✅ Solution : Fournir extraits de code réels à l'IA
```

**3. Faut CONNAÎTRE Divio pour bien prompter**
```
❌ "Documente mon projet" → résultat médiocre
✅ "Documente selon framework Divio" → résultat structuré
```

**4. IA ne connaît pas votre projet**
```
❌ Demander sans contexte → hallucinations
✅ Donner README + exemples de code → précis
```

### Best Practice

**Humain = Stratégie** (connaît Divio + métier)  
**IA = Exécution** (rapidité + cohérence)

---

## SLIDE 10 : Méthodologie - Process Reproductible

### Comment Faire (5 Étapes)

**1️⃣ BRIEF IA** (5 min)
```
"Documente projet X selon framework Divio"
+ Donner README actuel
+ Donner structure fichiers (ls -R ou tree)
```

**2️⃣ GÉNÉRER STRUCTURE** (5 min)
```
Valider arborescence avant contenu :
docs/tutorials/, docs/how-to/, docs/reference/, docs/explanation/
```

**3️⃣ GÉNÉRER CONTENU** (20-30 min)
```
Par catégorie, fichier par fichier :
"Génère tutorials/01-getting-started.md"
"Génère how-to/run-tests.md"
"Génère reference/api-reference.md"
"Génère explanation/why-multi-agent.md"
```

**4️⃣ RAFFINER** (10-20 min)
```
Itérations rapides :
"Supprime section X"
"Ajoute exemple Y"
"Ton plus concis dans Z"
"Corrige erreur dans W"
```

**5️⃣ GIT COMMIT** (5 min)
```
IA peut créer commits + messages :
"Commit tout avec message descriptif et push"
```

**⏱️ Total : ~1h pour documentation complète**

---

## SLIDE 11 : Call-to-Action

### 🚀 Essayez sur VOTRE Projet

**Prompt de Démarrage** (copier-coller)

```
Je veux documenter mon projet selon le framework Divio.

C'est un [type de projet : API REST / CLI tool / bibliothèque / app web].

Voici mon README actuel :
[coller votre README]

Crée la structure docs/ avec les 4 catégories Divio.
```

**Puis itérez** :
```
"Génère tutorials/01-getting-started.md pour débutants"
"Génère reference/api-reference.md avec tous les endpoints"
```

### Ressources

📚 **Framework Divio** : https://documentation.divio.com/  
💻 **Notre exemple** : https://github.com/elsasngr/ucp-merchant-test/tree/stageocto/docs  
🤖 **Outils IA** : Claude Code / ChatGPT / Cursor / GitHub Copilot

### ROI

⏱️ **Temps investi** : 1 heure  
📖 **Résultat** : Documentation professionnelle pour des années  
📈 **Impact** : Onboarding rapide, adoption élevée, moins de questions

---

## SLIDE 12 : Récapitulatif

### Ce Que Vous Avez Appris

✅ **Framework Divio** : 4 types de documentation (Tutorials / How-to / Reference / Explanation)

✅ **IA comme accélérateur** : Comprend, génère, maintient cohérence, itère rapidement

✅ **Process reproductible** : 5 étapes en ~1h

✅ **ROI mesuré** : 95% gain de temps, cohérence parfaite

✅ **Pièges évités** : Vérifier séparation types, fournir contexte, connaître Divio

### Prochaine Étape

**Testez ce soir** sur un de vos projets !

Commencez petit : 1 tutorial + 1 reference  
Puis étendez selon besoins

---

## Annexe : Prompts Efficaces

### Prompt Initial Complet

```
Je veux restructurer la documentation de mon projet selon le framework Divio.

CONTEXTE PROJET :
- Type : [API REST / CLI / bibliothèque / app web]
- Langage : [Go / Python / JavaScript / ...]
- Public : [développeurs / utilisateurs finaux / ops]

ÉTAT ACTUEL :
[Coller README ou décrire structure existante]

OBJECTIF :
Créer une documentation professionnelle organisée selon Divio :
- tutorials/ : guides apprentissage pas-à-pas
- how-to/ : solutions à problèmes spécifiques
- reference/ : specs techniques complètes
- explanation/ : contexte et décisions architecture

ÉTAPE 1 : Propose-moi d'abord la structure de dossiers/fichiers.
```

### Prompts de Génération

```
"Génère tutorials/01-getting-started.md :
- Durée : 15 minutes
- Public : débutant total
- Objectif : compiler, lancer, première requête API
- Format : étapes numérotées avec résultats attendus"
```

```
"Génère reference/api-reference.md :
- Tous les endpoints REST
- Pour chaque : method, path, params, response, exemple curl
- Format : tableau + exemples concrets"
```

### Prompts de Raffinement

```
"Le tutorial 1 est trop technique, simplifie pour un débutant"
```

```
"Ajoute un exemple concret avec code réel dans how-to/configure-X.md"
```

```
"Supprime toutes les sections 'à venir' des READMEs"
```

### Prompt Git

```
"Commit tous les fichiers docs/ avec un message descriptif et push sur origin/main"
```

---

## Annexe : Checklist Qualité

### ✅ Votre Documentation Est Bonne Si...

**Structure** :
- [ ] 4 dossiers séparés (tutorials, how-to, reference, explanation)
- [ ] README.md avec navigation claire
- [ ] Chaque dossier a son propre README.md

**Tutorials** :
- [ ] Pas-à-pas numérotés
- [ ] Résultat concret à la fin
- [ ] Durée estimée indiquée
- [ ] Prérequis clairs

**How-to** :
- [ ] Orienté résultat (résoudre problème X)
- [ ] Instructions concises
- [ ] Section "Troubleshooting"
- [ ] Pas d'explications longues

**Reference** :
- [ ] Exhaustif (tous les endpoints / fonctions / options)
- [ ] Format prévisible
- [ ] Exemples concrets copiables-collables
- [ ] Pas de narratif

**Explanation** :
- [ ] Répond au "pourquoi"
- [ ] Contexte et alternatives
- [ ] Trade-offs expliqués
- [ ] Liens vers ADRs si applicable

**Cohérence** :
- [ ] Même ton partout
- [ ] Pas de mélange de types
- [ ] Navigation facile (liens croisés)
- [ ] Git historique propre

---

## Contact & Questions

**Présentateur** : [Votre nom]  
**Projet exemple** : UCP Merchant Test  
**Repository** : https://github.com/elsasngr/ucp-merchant-test

**Questions ?**

---

*Présentation créée avec IA + Framework Divio*  
*Temps de création : 60 minutes*  
*Fichier source : Markdown (convertible Marp/Reveal.js/PDF)*
