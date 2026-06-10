# VALEURS DE TEST - SYSTÈME MULTI-AGENTS UNIFIÉ

## 🚀 Commande pour tout lancer

```bash
./run_unified_demo.sh
```

Cette commande lance :
1. **Shopping Graph** (port 9000) - Base de données des produits et concurrents
2. **Observability Hub** (port 9002) - Monitoring des agents
3. **Arena Merchant** (port 8888) - Interface marchande avec système multi-agents
4. **Client Agent** - Agent acheteur Gemini

## 🎯 Dashboard

**URL**: http://localhost:8888

Cliquez sur **"Test AUTO_COMPETE"** pour déclencher le système multi-agents.

## 🧪 SCÉNARIOS DE TEST

### ✅ Scénario 1 : Client PREMIUM (À GARDER ABSOLUMENT)

**Données à saisir** :
- **Customer ID**: `premium_vip_001`
- **Product**: `casque_bluetooth`
- **Code promo**: `AUTO_COMPETE`

**Profil client simulé** :
- Dépensé : $1,500+
- Achats : 15+
- Tier : PREMIUM

**Résultat attendu** :
```
👤 AGENT 2: CUSTOMER GROWTH
   ✅ OUI - Garder ce client
   Tier: premium
   Réduction suggérée: 15%
   
📊 AGENT 3: COMPÉTITIVITÉ
   Position marché: 2/5
   Prix compétitif: $57.00
   
🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix final: ~$48.45 (prix compétitif - 15%)
   Stratégie: vip_retention
   Marge: ~3-5%
```

---

### ✅ Scénario 2 : Client GOLD (Important)

**Données à saisir** :
- **Customer ID**: `gold_customer_002`
- **Product**: `laptop_pro`
- **Code promo**: `AUTO_COMPETE`

**Profil client simulé** :
- Dépensé : $500-$999
- Achats : 8-12
- Tier : GOLD

**Résultat attendu** :
```
👤 AGENT 2: CUSTOMER GROWTH
   ✅ OUI - Garder ce client
   Tier: gold
   Réduction suggérée: 10%
   
📊 AGENT 3: COMPÉTITIVITÉ
   Prix compétitif selon marché
   
🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix final: prix compétitif - 10%
   Stratégie: vip_retention
```

---

### ✅ Scénario 3 : Client SILVER (Bon client)

**Données à saisir** :
- **Customer ID**: `silver_customer_003`
- **Product**: `souris_gaming`
- **Code promo**: `AUTO_COMPETE`

**Profil client simulé** :
- Dépensé : $200-$499
- Achats : 4-7
- Tier : SILVER

**Résultat attendu** :
```
👤 AGENT 2: CUSTOMER GROWTH
   ✅ OUI - Garder ce client
   Tier: silver
   Réduction suggérée: 5%
   
📊 AGENT 3: COMPÉTITIVITÉ
   Prix compétitif selon marché
   
🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix final: prix compétitif - 5%
   Stratégie: vip_retention
```

---

### ❌ Scénario 4 : Client STANDARD (Pas prioritaire)

**Données à saisir** :
- **Customer ID**: `standard_customer_999`
- **Product**: `clavier_meca`
- **Code promo**: `AUTO_COMPETE`

**Profil client simulé** :
- Dépensé : <$200
- Achats : 1-2
- Tier : STANDARD

**Résultat attendu** :
```
👤 AGENT 2: CUSTOMER GROWTH
   ❌ NON - Pas prioritaire
   Tier: standard
   Réduction suggérée: 0%
   
📊 AGENT 3: COMPÉTITIVITÉ
   Prix compétitif selon marché
   
🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix final: prix compétitif (sans bonus)
   Stratégie: competitive_pricing
```

---

## 📊 Comprendre les Tiers Clients

| Tier     | Dépensé Total | Réduction | Priorité        |
|----------|---------------|-----------|-----------------|
| PREMIUM  | ≥ $1,000      | 15%       | ⭐⭐⭐⭐⭐ MAX    |
| GOLD     | $500 - $999   | 10%       | ⭐⭐⭐⭐ Élevée  |
| SILVER   | $200 - $499   | 5%        | ⭐⭐⭐ Moyenne   |
| STANDARD | < $200        | 0%        | ⭐ Standard      |

## 🔍 Observer les Agents en Action

Le code promo **AUTO_COMPETE** déclenche le flux suivant :

```
┌─────────────────┐
│ Acheteur Lambda │
│   demande prix  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  AGENT 1: VENDEUR (Orchestrateur)   │
│  "Quel prix donner à ce client ?"   │
└──────┬──────────────────────┬───────┘
       │                      │
       ▼                      ▼
┌──────────────────┐  ┌──────────────────────┐
│ AGENT 2:         │  │ AGENT 3:             │
│ CUSTOMER GROWTH  │  │ COMPÉTITIVITÉ        │
│                  │  │                      │
│ Garder client ?  │  │ Prix compétitif ?    │
│ → OUI/NON        │  │ → $XX.XX             │
│ → Tier           │  │ → Position marché    │
│ → % réduction    │  │ → Stratégie          │
└──────┬───────────┘  └──────┬───────────────┘
       │                     │
       └──────────┬──────────┘
                  ▼
       ┌─────────────────────┐
       │  AGENT 1: VENDEUR   │
       │  (Synthèse)         │
       │                     │
       │  Prix compétitif    │
       │  - Bonus VIP        │
       │  = Prix final       │
       └──────────┬──────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Acheteur Lambda│
         │  reçoit prix   │
         └────────────────┘
```

## 📝 Logs utiles

```bash
# Voir tous les logs en temps réel
tail -f /tmp/arena.log

# Voir uniquement les décisions des agents
tail -f /tmp/arena.log | grep "Agent"

# Voir le Shopping Graph
tail -f /tmp/shopping-graph.log

# Voir l'Observability Hub
tail -f /tmp/obs-hub.log
```

## 🎮 Test Standalone (sans UI)

Si tu veux juste tester le système multi-agents sans l'interface :

```bash
go run ./pkg/pricing-unified/example/main.go
```

Ce script teste directement avec :
- Un client premium (1500$ dépensés)
- Un client standard (100$ dépensés)

## 🛑 Arrêter tous les services

Appuie sur **Ctrl+C** dans le terminal où tu as lancé `run_unified_demo.sh`.

Ou manuellement :
```bash
pkill -f "bin/shopping-graph"
pkill -f "bin/obs-hub"
pkill -f "bin/arena"
pkill -f "bin/client-agent"
```

## ⚙️ Configuration Google Cloud

Le script configure automatiquement :
```bash
export GOOGLE_CLOUD_PROJECT="bsjxygz-gcp-octo-lille"
```

Si tu as des erreurs avec le Client Agent, vérifie que ton projet Google Cloud est bien configuré.

## 🔧 Troubleshooting

**Problème**: "Port already in use"
```bash
# Trouve et tue le processus
lsof -ti:9000 | xargs kill -9  # Shopping Graph
lsof -ti:9002 | xargs kill -9  # Obs Hub
lsof -ti:8888 | xargs kill -9  # Arena
```

**Problème**: "Shopping Graph connection refused"
```bash
# Vérifie que le Shopping Graph tourne
curl http://localhost:9000/health
```

**Problème**: Dashboard ne charge pas
```bash
# Rebuild Arena
cd demo
go build -o bin/arena ./cmd/arena
./bin/arena --port 8888
```
