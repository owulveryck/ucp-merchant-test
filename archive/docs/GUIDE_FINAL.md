# 🎮 GUIDE COMPLET - SYSTÈME MULTI-AGENTS + ARÈNE

## 🚀 LANCEMENT

```bash
./demo.sh
```

Attends ~15 secondes que tout démarre.

---

## 📋 ÉTAPES POUR JOUER

### 1️⃣ Crée tes marchands

Va sur : **http://localhost:8888**

Crée **2-3 marchands** avec des noms différents (ex: MarchandA, MarchandB, etc.)

---

### 2️⃣ Configure le prix de chaque marchand

Pour **CHAQUE marchand** :

1. Clique sur le marchand → Dashboard
2. Clique sur **"💡 Calculer le meilleur prix"**
3. **Regarde l'analyse des 3 agents** :
   - 👤 **Agent 2** : Customer Growth (client Gold/Premium/Silver ?)
   - 📊 **Agent 3** : Compétitivité (position marché)
   - 🎯 **Agent 1** : Vendeur (décision finale)
4. Clique sur **"✨ Appliquer ce prix"**

✅ Ton marchand a maintenant le prix optimal calculé par les agents !

---

### 3️⃣ Ouvre l'arène

Va sur : **http://localhost:9002/arena**

Tu vas voir :
- 📊 Tous les marchands en compétition
- 💰 Leurs prix
- 🏆 Qui est le moins cher

---

### 4️⃣ Lance l'agent acheteur

**Option A : DIRECTEMENT DANS L'ARÈNE** (recommandé !)

Dans l'arène (http://localhost:9002/arena), tape ce que tu veux acheter :
- "Achète un casque pas cher"
- "Je veux un laptop rapide"
- "Trouve-moi une souris gaming"

Clique **"🛒 Acheter"** ou appuie sur **Entrée**

**Option B : Script shell**

Ouvre un **NOUVEAU terminal** et tape :

```bash
./acheter.sh
```

Choisis ce que tu veux acheter :
1. Casque audio (budget $100)
2. Laptop (budget $1000)
3. Souris gaming (budget $80)
4. Autre (personnalisé)

L'agent acheteur va :
1. 🔍 Chercher dans le Shopping Graph
2. 📊 Comparer les prix de tous les marchands
3. 🛒 Acheter chez le moins cher

---

## 🏆 POURQUOI TU VAS GAGNER ?

Le système **3-agents** calcule automatiquement :

### Agent 2 : Customer Growth
- Détecte le tier du client (Gold par défaut)
- Applique un **bonus fidélité de 10%**

### Agent 3 : Compétitivité  
- Compare avec tous les concurrents
- Trouve leur prix le plus bas
- Calcule comment les battre

### Agent 1 : Vendeur
- Synthétise tout
- Prix compétitif **MOINS** le bonus fidélité
- **Tu deviens le moins cher !** 🎉

---

## 📊 EXEMPLE CONCRET

### Situation de départ :
- MarchandA : **$72.56**
- MarchandB : **$65.00**
- Toi : **$70.00**

### Après "Calculer meilleur prix" :
- Agent 3 voit : "Concurrent le moins cher = $65.00"
- Agent 3 recommande : **$63.75** (pour battre $65)
- Agent 2 dit : "Client Gold, -10%"
- Agent 1 décide : **$57.38** (63.75 - 10%)

### Résultat final :
- MarchandA : $72.56
- MarchandB : $65.00
- **Toi : $57.38** 🏆 ← **TU GAGNES !**

L'agent acheteur achète chez toi !

---

## 🎯 COMMANDES UTILES

### Relancer tout :
```bash
./demo.sh
```

### Envoyer une commande d'achat :
```bash
./acheter.sh
```

### Voir les logs des agents :
```bash
tail -f logs/arena.log | grep -E "Agent|Vendeur|Customer|Competitiveness"
```

### Arrêter tout :
```
Ctrl+C (dans le terminal de ./demo.sh)
```

---

## 🌐 URLS IMPORTANTES

| Service | URL | Description |
|---------|-----|-------------|
| **Dashboard** | http://localhost:8888 | Créer/gérer les marchands |
| **Arène** | http://localhost:9002/arena | Voir la compétition en temps réel |
| **Shopping Graph** | http://localhost:9000 | Base de données produits |

---

## 🤖 LE SYSTÈME 3-AGENTS EN DÉTAIL

### Flux de décision :

```
1. Tu cliques "Calculer meilleur prix"
   ↓
2. AGENT 1 (Vendeur) coordonne
   ↓
3. ┌─ AGENT 2 (Customer Growth)
   │  "Client Gold → -10%"
   │
   └─ AGENT 3 (Compétitivité)
      "Concurrent = $65 → recommande $63.75"
   ↓
4. AGENT 1 synthétise
   "$63.75 - 10% = $57.38"
   ↓
5. Tu appliques ce prix
   ↓
6. Tu gagnes ! 🏆
```

---

## 💡 ASTUCES

### Pour toujours gagner :
1. Laisse les autres marchands avec leurs prix manuels
2. Utilise "Calculer meilleur prix" sur TON marchand
3. Applique le prix recommandé
4. Tu seras TOUJOURS le moins cher !

### Pour voir le raisonnement complet :
Après "Calculer meilleur prix", regarde l'affichage détaillé :
- Chaque agent montre son analyse
- Tu vois exactement pourquoi ce prix a été choisi
- Tout est transparent !

---

## 🎉 C'EST PARTI !

1. `./demo.sh` → Lance tout
2. Crée 2-3 marchands
3. Configure leurs prix avec les agents
4. Regarde l'arène : http://localhost:9002/arena
5. `./acheter.sh` → Lance l'acheteur
6. **GAGNE !** 🏆
