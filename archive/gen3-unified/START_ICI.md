# ⚡ COMMENCE ICI - GUIDE ULTRA-SIMPLE

## 🎯 Ce que tu vas faire

Tu vas tester le système multi-agents qui décide automatiquement du meilleur prix pour chaque client.

---

## 🚀 ÉTAPE 1 : Lance tout

Dans ton terminal :

```bash
./run_unified_demo.sh
```

Attends ~10 secondes. Tu verras :
```
✓ Shopping Graph started
✓ Observability Hub started  
✓ Arena Merchant started
```

**NE ferme PAS ce terminal** (laisse-le tourner).

---

## 🌐 ÉTAPE 2 : Ouvre le dashboard

Ouvre ton navigateur sur :

```
http://localhost:8888
```

Tu verras une liste de marchands. **Clique sur n'importe quel marchand** (par exemple "marchandA").

---

## ⚙️ ÉTAPE 3 : Active le système multi-agents

Sur le dashboard du marchand :

1. Cherche la section **"ALGORITHMES PRIX"**
2. Clique sur le bouton **"COMPÉTITIF"**

Le bouton devient rouge quand c'est activé.

💡 C'est ce bouton qui active le système multi-agents !

---

## 🧪 ÉTAPE 4 : Teste le système

**Option A : Test rapide (sans UI)**

Ouvre un NOUVEAU terminal et lance :

```bash
./test_quick.sh
```

Tu verras les 3 agents travailler ensemble :
```
👤 AGENT 2: CUSTOMER GROWTH
   ✅ OUI - Garder ce client
   Tier: premium
   Réduction suggérée: 15%

📊 AGENT 3: COMPÉTITIVITÉ
   Position marché: 2/5
   Prix recommandé: $57.00

🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix final: $48.45
   Stratégie: vip_retention
```

**Option B : Via le dashboard**

Sur le dashboard (http://localhost:8888), tu peux :
- Modifier le prix avec le slider
- Voir les changements en temps réel
- L'algo "COMPÉTITIF" utilise le système multi-agents

---

## 📊 Comprendre les résultats

Quand le système multi-agents est activé :

**AGENT 2 (Customer Growth)** analyse le client :
- Client PREMIUM → -15% de réduction
- Client GOLD → -10% de réduction
- Client SILVER → -5% de réduction
- Client STANDARD → Pas de réduction

**AGENT 3 (Compétitivité)** analyse le marché :
- Compare avec les concurrents
- Trouve le meilleur prix
- Vérifie qu'on reste compétitif

**AGENT 1 (Vendeur)** décide :
- Prend le prix compétitif
- Applique le bonus VIP si client premium
- S'assure de ne pas perdre d'argent

---

## 🎮 Clients de test

Si tu veux tester avec différents profils clients via l'API :

```bash
# Client PREMIUM (réduction 15%)
curl -X POST http://localhost:8888/{tenant}/checkout \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "premium_vip_001", "discount_codes": ["AUTO_COMPETE"]}'

# Client STANDARD (pas de réduction)
curl -X POST http://localhost:8888/{tenant}/checkout \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "standard_customer_999", "discount_codes": ["AUTO_COMPETE"]}'
```

Remplace `{tenant}` par le tenant ID (par exemple `marchandA`).

---

## 🛑 Arrêter

Dans le terminal où tu as lancé `./run_unified_demo.sh`, appuie sur :

```
Ctrl+C
```

Tous les services s'arrêtent automatiquement.

---

## ❓ Problèmes ?

**Le dashboard ne charge pas** :
- Attends 15 secondes après le lancement
- Vérifie que le port 8888 est libre : `lsof -ti:8888`

**Je ne vois pas le bouton COMPÉTITIF** :
- Assure-toi d'avoir cliqué sur un marchand (pas juste la page d'accueil)
- Cherche dans la section "ALGORITHMES PRIX"

**Test rapide ne marche pas** :
- Lance d'abord `./run_unified_demo.sh`
- Puis dans un autre terminal : `./test_quick.sh`

---

## 📚 Plus d'infos ?

- **COMMENT_TESTER.md** : Guide détaillé des tests
- **TEST_VALUES.md** : Tous les scénarios
- **COMMANDES.txt** : Toutes les commandes

---

## ⚡ TL;DR

```bash
# 1. Lance tout
./run_unified_demo.sh

# 2. Ouvre le navigateur
http://localhost:8888

# 3. Clique sur un marchand
# 4. Active l'algo "COMPÉTITIF"

# 5. Test rapide (autre terminal)
./test_quick.sh

# 6. Arrête tout
Ctrl+C (dans le terminal du step 1)
```

Le système multi-agents fonctionne quand l'algo **COMPÉTITIF** est activé !
