# 🎮 Démo Multi-Agents - Instructions

## 🚀 Lancer la Démo

```bash
./run_arena_demo.sh
```

Vous verrez :

```
✅ Tous les services sont lancés !

┌──────────────────────────────────────────────────────────┐
│  🌐 Ouvrez dans votre navigateur:                        │
│     http://localhost:8080/                               │
│                                                           │
│  📝 ÉTAPES:                                              │
│     1. Créez 2-3 marchands                               │
│     2. Configurez des prix différents                    │
│     3. Testez AUTO_COMPETE avec un checkout              │
│                                                           │
│  🤖 Pour tester AUTO_COMPETE :                           │
│     - Dans l'interface Arena, créez un checkout          │
│     - Utilisez le code promo: AUTO_COMPETE               │
│     - Regardez le prix s'ajuster automatiquement !       │
│                                                           │
│  🛑 Pour arrêter: Ctrl+C                                 │
└──────────────────────────────────────────────────────────┘

Appuyez sur Ctrl+C pour arrêter tous les services
```

---

## 🎯 Utilisation de l'Interface Arena

### 1. Créer des Merchants

1. Ouvrez **http://localhost:8080**
2. Cliquez sur **"Créer un nouveau marchand"**
3. Créez 2-3 merchants avec des noms différents :
   - SuperShop
   - MegaMart
   - BudgetBuy

### 2. Configurer les Prix

Pour chaque merchant :
1. Cliquez sur son dashboard
2. Utilisez le **slider de prix** pour ajuster
3. Définissez des **codes promo** (SAVE10, WELCOME20...)
4. Configurez les **options de livraison**

**Exemple de configuration :**
- SuperShop : Roses à $65.00
- MegaMart : Roses à $59.99 ← **Le moins cher**
- BudgetBuy : Roses à $70.00

### 3. Tester AUTO_COMPETE

**Option A : Interface Arena**
1. Dans un dashboard merchant (ex: SuperShop)
2. Créez un checkout avec un produit
3. Appliquez le code **AUTO_COMPETE**
4. 🎉 Le prix s'ajuste automatiquement pour battre MegaMart !

**Option B : Script de test**
```bash
./test_auto_compete.sh
```

Ce script va :
- Créer un checkout automatiquement
- Appliquer AUTO_COMPETE
- Afficher le prix avant/après

---

## 🤖 Les 4 Agents en Action

Quand vous appliquez AUTO_COMPETE, regardez les logs :

```bash
tail -f logs/arena.log
```

Vous verrez :

```
[Orchestrator] Starting competitive pricing analysis...

[Agent 1 - Price Intelligence]
[Orchestrator] Price Intelligence: rank 2/3, lowest: $59.99 (MegaMart)

[Agent 2 - Market Analysis]
[Orchestrator] Market Analysis: follower position, stable trend

[Agent 3 - Strategy Recommender]
[Orchestrator] Strategy: balanced, target: $56.99, confidence: 80%
[Orchestrator] Reasoning: ["Standard competitive positioning"]

[Agent 4 - Margin Validator]
[Orchestrator] ✅ Pricing approved: $56.99 (discount: $9.24, margin: 25%)
```

---

## 📊 Intelligence Compétitive

Dans le dashboard de chaque merchant :
- **Onglet "Competitive Intel"**
- Voir les prix concurrents en temps réel
- Recommandations de prix
- Bouton "Appliquer ce prix"

---

## 🧪 Scénarios de Test

### Scénario 1 : Stock Bas → Stratégie Aggressive

1. Dans l'Arena, baissez le stock d'un produit à 10 unités
2. Appliquez AUTO_COMPETE
3. **Résultat** : Discount plus fort (10% au lieu de 5%)

### Scénario 2 : Déjà Leader → Stratégie Premium

1. Configurez un merchant pour être le moins cher
2. Appliquez AUTO_COMPETE
3. **Résultat** : Pas de discount (garde sa marge)

### Scénario 3 : Guerre des Prix → Match

1. Baissez progressivement les prix de tous les merchants
2. Appliquez AUTO_COMPETE
3. **Résultat** : Match le prix exact du concurrent

---

## 🛑 Arrêter la Démo

Dans le terminal où tourne `run_arena_demo.sh` :

```
Ctrl+C
```

Ou depuis un autre terminal :

```bash
./stop_demo.sh
```

---

## 📝 Fichiers de Logs

```
logs/
├── shopping-graph.log   # Shopping Graph
└── arena.log            # Arena + Merchants (VOIR ICI pour les agents!)
```

---

## ❓ Problèmes Fréquents

### "Port already in use"
```bash
./stop_demo.sh
./run_arena_demo.sh
```

### "AUTO_COMPETE n'applique pas de discount"
Raisons possibles :
- Vous êtes déjà le moins cher
- Pas assez de merchants concurrents (créez-en 2-3)
- Contraintes de marge

### "Aucun merchant trouvé"
Créez des merchants dans l'interface Arena : http://localhost:8080

---

## 🎓 En Résumé

1. **Lance** : `./run_arena_demo.sh`
2. **Ouvre** : http://localhost:8080
3. **Crée** : 2-3 merchants avec prix différents
4. **Teste** : Code AUTO_COMPETE
5. **Regarde** : `tail -f logs/arena.log`
6. **Arrête** : Ctrl+C

**C'est tout !** 🎉
