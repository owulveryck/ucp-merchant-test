# 🧪 COMMENT TESTER LE SYSTÈME MULTI-AGENTS

## ⚠️ Important : Comment déclencher les agents

Le système multi-agents se déclenche automatiquement quand tu utilises le code promo **`AUTO_COMPETE`** lors d'un checkout.

Il n'y a **pas de bouton "Test AUTO_COMPETE"** sur le dashboard. C'est le code promo qui active le système.

---

## 🚀 MÉTHODE 1 : Via l'API (Recommandé pour tester)

### Étape 1 : Lance la démo

```bash
./run_unified_demo.sh
```

Attends que tous les services démarrent (~10 secondes).

### Étape 2 : Teste avec curl

```bash
# Test rapide du système multi-agents
curl -X POST http://localhost:8888/api/test-auto-compete
```

Cette commande déclenche le système multi-agents et affiche la décision complète des 4 agents (système existant).

---

## 🌐 MÉTHODE 2 : Via le Dashboard (Simulation complète)

### Étape 1 : Ouvre le dashboard

```
http://localhost:8888
```

### Étape 2 : Crée un checkout avec le code promo AUTO_COMPETE

Tu dois utiliser l'API UCP pour créer un checkout. Voici comment :

#### 2.1 : Crée un panier (cart)

```bash
curl -X POST http://localhost:8888/carts \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": "casque_audio",
        "quantity": 1
      }
    ]
  }'
```

Récupère le `cart_id` de la réponse.

#### 2.2 : Crée un checkout avec AUTO_COMPETE

```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "cart_id": "le_cart_id_récupéré",
    "customer_id": "premium_vip_001",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Le code `AUTO_COMPETE` dans `discount_codes` déclenche le système multi-agents !**

---

## 📊 Ce qui se passe quand AUTO_COMPETE est utilisé

```
1. Le code AUTO_COMPETE est détecté
   ↓
2. AGENT 1 (Vendeur) est appelé
   ↓
3. AGENT 1 consulte AGENT 2 (Customer Growth)
   "Client premium_vip_001 - Garder ?"
   ↓
4. AGENT 2 analyse et répond
   "OUI, tier PREMIUM, -15%"
   ↓
5. AGENT 1 consulte AGENT 3 (Compétitivité)
   "Produit casque_audio - Prix compétitif ?"
   ↓
6. AGENT 3 analyse le marché et répond
   "Position 2/5, prix recommandé $57"
   ↓
7. AGENT 1 synthétise
   "$57 (compétitif) - 15% (VIP) = $48.45"
   ↓
8. Le client reçoit le prix final
```

---

## 🧪 Tests rapides sans UI

### Test 1 : Système multi-agents standalone

```bash
./test_quick.sh
```

Affiche directement les 3 agents en action avec 2 scénarios :
- Client PREMIUM → Réduction 15%
- Client STANDARD → Pas de réduction

### Test 2 : Via l'API de test

```bash
# Lance la démo
./run_unified_demo.sh

# Dans un autre terminal
curl -X POST http://localhost:8888/test-auto-compete
```

---

## 📋 Scénarios de test détaillés

### Scénario 1 : Client PREMIUM

```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "casque_audio", "quantity": 1}],
    "customer_id": "premium_vip_001",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Résultat attendu** :
- Agent 2 : OUI garder, tier PREMIUM, -15%
- Agent 3 : Prix compétitif calculé
- Agent 1 : Prix final = prix compétitif - 15%

### Scénario 2 : Client GOLD

```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "laptop_pro", "quantity": 1}],
    "customer_id": "gold_customer_002",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Résultat attendu** :
- Agent 2 : OUI garder, tier GOLD, -10%
- Agent 3 : Prix compétitif calculé
- Agent 1 : Prix final = prix compétitif - 10%

### Scénario 3 : Client SILVER

```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "souris_gaming", "quantity": 1}],
    "customer_id": "silver_customer_003",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Résultat attendu** :
- Agent 2 : OUI garder, tier SILVER, -5%
- Agent 3 : Prix compétitif calculé
- Agent 1 : Prix final = prix compétitif - 5%

### Scénario 4 : Client STANDARD

```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "clavier_meca", "quantity": 1}],
    "customer_id": "standard_customer_999",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Résultat attendu** :
- Agent 2 : NON garder, tier STANDARD, 0%
- Agent 3 : Prix compétitif calculé
- Agent 1 : Prix final = prix compétitif (sans bonus)

---

## 🔍 Voir les logs des agents

```bash
# Voir toutes les décisions en temps réel
tail -f /tmp/arena.log | grep "Agent"
```

Tu verras :
```
[Agent Vendeur] Demande de prix pour casque_audio
[Agent Vendeur] → Consultation Agent 2 (Customer Growth)
[Agent Customer Growth] Analyzing customer: premium_vip_001
[Agent Customer Growth] Decision: ShouldRetain=true, Tier=premium
[Agent Vendeur] → Consultation Agent 3 (Compétitivité)
[Agent Compétitivité] Analyzing competitiveness...
[Agent Vendeur] ✓ Prix final décidé: $48.45
```

---

## 🎯 TL;DR

**Pour tester rapidement** :
```bash
./test_quick.sh
```

**Pour tester via l'API** :
```bash
./run_unified_demo.sh
# Dans un autre terminal :
curl -X POST http://localhost:8888/api/test-auto-compete
```

**Pour tester un scénario complet** :
```bash
curl -X POST http://localhost:8888/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "casque_audio", "quantity": 1}],
    "customer_id": "premium_vip_001",
    "discount_codes": ["AUTO_COMPETE"]
  }'
```

**Le code magique** : `AUTO_COMPETE` dans `discount_codes`

**Voir les logs** : `tail -f /tmp/arena.log | grep "Agent"`
