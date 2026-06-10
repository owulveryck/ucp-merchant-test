# 🚀 DÉMARRAGE RAPIDE - SYSTÈME MULTI-AGENTS

## Option 1 : Test Rapide (2 secondes)

```bash
./test_quick.sh
```

**Ce que tu vois :**
- ✅ Client PREMIUM → Prix $60 → Final $51 (réduction 15%)
- ❌ Client STANDARD → Prix $60 → Final $60 (pas de réduction)

**Idéal pour** : Vérifier que le système fonctionne

---

## Option 2 : Démo Complète (avec interface)

```bash
./run_unified_demo.sh
```

**Ce qui se lance :**
1. Shopping Graph (port 9000)
2. Observability Hub (port 9002)
3. Arena Dashboard (port 8888) ← **L'interface**
4. Client Agent (Gemini)

**Puis ouvre :** http://localhost:8888

---

## 🎯 Sur le Dashboard

Clique sur **"Test AUTO_COMPETE"** pour déclencher le système multi-agents.

---

## 📋 Valeurs de Test

### Pour tester les différents tiers clients :

| Customer ID            | Tier     | Réduction | Résultat                    |
|------------------------|----------|-----------|------------------------------|
| `premium_vip_001`      | PREMIUM  | 15%       | Prix compétitif - 15%       |
| `gold_customer_002`    | GOLD     | 10%       | Prix compétitif - 10%       |
| `silver_customer_003`  | SILVER   | 5%        | Prix compétitif - 5%        |
| `standard_customer_999`| STANDARD | 0%        | Prix compétitif (aucun bonus)|

### Produits à tester :
- casque_bluetooth
- laptop_pro
- souris_gaming
- clavier_meca

### Code promo magique :
**AUTO_COMPETE** ← Déclenche le système multi-agents

---

## 📊 Comment ça marche ?

```
Acheteur demande un prix
         ↓
    AGENT 1 (VENDEUR)
    "Quel prix donner ?"
         ↓
    ┌────┴────┐
    ↓         ↓
AGENT 2    AGENT 3
Customer   Compétitivité
Growth
    ↓         ↓
"Garder?"  "Prix
OUI/NON    compétitif?"
    ↓         ↓
    └────┬────┘
         ↓
    AGENT 1
    Synthèse
         ↓
    Prix final
```

### AGENT 2 : Customer Growth
- Analyse le client (historique, dépenses)
- Décide s'il faut le garder : OUI/NON
- Propose une réduction selon le tier

### AGENT 3 : Compétitivité
- Analyse le marché (prix concurrents)
- Vérifie notre position
- Propose un prix compétitif

### AGENT 1 : Vendeur
- Reçoit les 2 analyses
- Fait la synthèse
- Décide du prix final

---

## 🛑 Arrêter

**Option 1 (démo complète)** : Appuie sur `Ctrl+C` dans le terminal

**Option 2 (manuellement)** :
```bash
pkill -f "bin/shopping-graph"
pkill -f "bin/obs-hub"
pkill -f "bin/arena"
pkill -f "bin/client"
```

---

## 📖 Documentation complète

- **TEST_VALUES.md** : Tous les scénarios de test
- **pkg/pricing-unified/README.md** : Documentation technique
- **run_unified_demo.sh** : Le script complet

---

## 🔧 Problèmes courants

**"Port already in use"** :
```bash
lsof -ti:9000 | xargs kill -9
lsof -ti:8888 | xargs kill -9
```

**"Shopping Graph connection refused"** :
C'est normal si tu utilises `test_quick.sh`. Le système fonctionne en mode dégradé.

Pour avoir le Shopping Graph, utilise `run_unified_demo.sh`.

---

## ⚡ TL;DR

**Test le plus rapide** :
```bash
./test_quick.sh
```

**Démo complète avec UI** :
```bash
./run_unified_demo.sh
# Puis va sur http://localhost:8888
# Clique "Test AUTO_COMPETE"
```

**Customer IDs à tester** :
- `premium_vip_001` (réduction 15%)
- `gold_customer_002` (réduction 10%)
- `silver_customer_003` (réduction 5%)
- `standard_customer_999` (pas de réduction)

**Code promo** : `AUTO_COMPETE`
