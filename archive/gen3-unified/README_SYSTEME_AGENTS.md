# 🤖 Système Multi-Agents - Documentation Complète

## 🚀 Démarrage rapide

**Commence ici** : [START_ICI.md](START_ICI.md)

Guide ultra-simple en 6 étapes pour lancer et tester le système.

---

## 📖 Documentation disponible

### 🎯 Pour démarrer
- **[START_ICI.md](START_ICI.md)** ← **À LIRE EN PREMIER**
- **[GUIDE_RAPIDE.txt](GUIDE_RAPIDE.txt)** - Guide visuel étape par étape
- **[COMMANDES.txt](COMMANDES.txt)** - Toutes les commandes utiles

### 🧪 Pour tester
- **[COMMENT_TESTER.md](COMMENT_TESTER.md)** - Guide détaillé des tests
- **[TEST_VALUES.md](TEST_VALUES.md)** - Scénarios et valeurs de test
- **[DEMARRAGE.md](DEMARRAGE.md)** - Options de lancement

### 📚 Pour comprendre
- **[README_MULTIAGENT.md](README_MULTIAGENT.md)** - Vue d'ensemble du système
- **[pkg/pricing-unified/README.md](pkg/pricing-unified/README.md)** - Documentation technique

---

## ⚡ Commandes principales

### Lancer la démo
```bash
./run_unified_demo.sh
```
Lance tous les services, puis ouvre http://localhost:8888

### Test rapide (sans UI)
```bash
./test_quick.sh
```
Montre les 3 agents en action en 2 secondes

### Arrêter
```
Ctrl+C
```
Dans le terminal où tu as lancé `run_unified_demo.sh`

---

## 🤖 Les 3 Agents

### AGENT 1 : VENDEUR (Orchestrateur)
**Question** : "Quel prix donner à ce client ?"  
**Rôle** : Coordonne les agents 2 et 3, décide du prix final

### AGENT 2 : CUSTOMER GROWTH
**Question** : "Est-ce un client à garder ?"  
**Rôle** : Analyse la valeur du client, propose une réduction

Tiers clients :
- 🌟 PREMIUM (≥$1000) : -15%
- 🥇 GOLD ($500-999) : -10%
- 🥈 SILVER ($200-499) : -5%
- ⚪ STANDARD (<$200) : 0%

### AGENT 3 : COMPÉTITIVITÉ
**Question** : "Sommes-nous compétitifs ?"  
**Rôle** : Analyse le marché, recommande un prix compétitif

---

## 🎯 Comment ça fonctionne ?

```
Client demande un prix
         ↓
    AGENT 1 (VENDEUR)
         ↓
    ┌────┴────┐
    ↓         ↓
AGENT 2    AGENT 3
(Client)   (Marché)
    ↓         ↓
    └────┬────┘
         ↓
  Prix final optimisé
```

**Exemple** :
1. Client PREMIUM demande un casque à $60
2. AGENT 2 : "Garder ce client, -15%"
3. AGENT 3 : "Prix compétitif = $57"
4. AGENT 1 : "$57 - 15% = **$48.45**"

---

## 🧪 Tester

### Option 1 : Test ultra-rapide
```bash
./test_quick.sh
```

### Option 2 : Via le dashboard
```bash
./run_unified_demo.sh
# Puis : http://localhost:8888
# Clique sur un marchand
# Active l'algo "COMPÉTITIF"
```

### Option 3 : Via l'API
```bash
./run_unified_demo.sh

# Dans un autre terminal :
# (Le système se déclenche avec le code promo AUTO_COMPETE)
```

Voir [COMMENT_TESTER.md](COMMENT_TESTER.md) pour tous les détails.

---

## 📁 Structure du projet

```
ucp-merchant-test/
├── START_ICI.md                    ← Commence ici !
├── GUIDE_RAPIDE.txt                ← Guide visuel
├── COMMENT_TESTER.md               ← Tests détaillés
├── TEST_VALUES.md                  ← Scénarios
├── run_unified_demo.sh             ← Lance tout
├── test_quick.sh                   ← Test rapide
└── pkg/pricing-unified/            ← Code du système
    ├── orchestrator.go             - Agent 1 (Vendeur)
    ├── agents/
    │   ├── customer_growth.go      - Agent 2
    │   └── competitiveness.go      - Agent 3
    └── models/types.go             - Types de données
```

---

## 🔧 Troubleshooting

### Services ne démarrent pas
```bash
# Vérifie les ports
lsof -ti:9000 | xargs kill -9  # Shopping Graph
lsof -ti:9002 | xargs kill -9  # Obs Hub
lsof -ti:8888 | xargs kill -9  # Arena
```

### Dashboard 404
```bash
# Attends 15 secondes après le lancement
# Vérifie les logs
tail -f /tmp/arena.log
```

### Voir les logs des agents
```bash
tail -f /tmp/arena.log | grep "Agent"
```

---

## ✅ Résumé

1. **Lance** : `./run_unified_demo.sh`
2. **Ouvre** : http://localhost:8888
3. **Active** : Bouton "COMPÉTITIF" sur le dashboard
4. **Teste** : `./test_quick.sh` (dans un autre terminal)
5. **Arrête** : `Ctrl+C`

Le système multi-agents ajuste automatiquement les prix selon :
- La valeur du client (Agent 2)
- La compétitivité du marché (Agent 3)
- Une synthèse optimale (Agent 1)

---

## 📞 Questions ?

- Guide simple : [START_ICI.md](START_ICI.md)
- Tests : [COMMENT_TESTER.md](COMMENT_TESTER.md)
- Commandes : [COMMANDES.txt](COMMANDES.txt)
- Vue d'ensemble : [README_MULTIAGENT.md](README_MULTIAGENT.md)

**Premier lancement ?** → Lis [START_ICI.md](START_ICI.md)
