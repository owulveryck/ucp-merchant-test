# 🎮 SYSTÈME MULTI-AGENTS - PAR OÙ COMMENCER ?

## 🚀 DEUX DÉMOS DISPONIBLES

### 1️⃣ **./demo.sh** - Mode Normal
**Pour qui** : Découvrir le système, avoir le contrôle total  
**Temps** : 5-10 minutes  
**Tu fais** : Créer manuellement 2-3 marchands, les configurer  

```bash
./demo.sh
```

📖 **Guide** : [GUIDE_FINAL.md](GUIDE_FINAL.md)

---

### 2️⃣ **./arena_challenge.sh** - Mode Challenge  
**Pour qui** : Expérience "wow", scénario David vs Goliath  
**Temps** : 2 minutes  
**Pré-configuré** : 4 concurrents agressifs déjà en place  

```bash
./arena_challenge.sh
```

📖 **Guide** : [CHALLENGE.md](CHALLENGE.md)

---

## 🎯 TU VEUX QUOI ?

### "Je veux comprendre comment ça marche"
→ Lance `./demo.sh`  
→ Lis [GUIDE_FINAL.md](GUIDE_FINAL.md)  
→ Crée tes marchands un par un  
→ Regarde l'analyse des 3 agents  

### "Je veux être impressionné rapidement"
→ Lance `./arena_challenge.sh`  
→ Lis [CHALLENGE.md](CHALLENGE.md)  
→ Crée TON marchand avec un prix élevé → tu perds  
→ Active le système 3-agents → tu gagnes !  

### "Je veux juste taper des commandes d'achat"
→ Lance `./demo.sh` OU `./arena_challenge.sh`  
→ Ouvre http://localhost:9002/arena  
→ Tape "Achète un casque pas cher"  
→ Lis [ARENA_INPUT.md](ARENA_INPUT.md)  

---

## ⚡ QUICK START (30 secondes)

```bash
# 1. Lance la démo
./demo.sh

# 2. Ouvre ton navigateur
# http://localhost:8888          <- Créer marchands
# http://localhost:9002/arena    <- Arène de compétition

# 3. Tape dans l'arène
"Achète un casque"

# 4. Pour arrêter
Ctrl+C
```

---

## 🤖 C'EST QUOI LE SYSTÈME 3-AGENTS ?

### Agent 1 : Vendeur (Orchestrateur)
Coordonne tout, décision finale

### Agent 2 : Customer Growth
Analyse le client, calcule bonus fidélité (0-15%)

### Agent 3 : Compétitivité
Espionne la concurrence, trouve le prix gagnant

**Résultat** : Prix optimal automatique qui bat les concurrents !

---

## 📚 TOUTE LA DOCUMENTATION

| Fichier | Pour quoi ? |
|---------|-------------|
| **START_HERE.md** | Ce fichier (vue d'ensemble) |
| **GUIDE_FINAL.md** | Mode normal complet |
| **CHALLENGE.md** | Mode challenge complet |
| **ARENA_INPUT.md** | Acheter dans l'arène |

---

## 🌐 URLS

Une fois lancé :

| URL | Quoi faire ? |
|-----|--------------|
| http://localhost:8888 | Créer/gérer tes marchands |
| http://localhost:9002/arena | Arène de compétition + input achat |
| http://localhost:9000 | Shopping Graph API |

---

## 💡 EN UN MOT

**./demo.sh** = Tu contrôles tout, parfait pour apprendre  
**./arena_challenge.sh** = 4 concurrents t'attendent, scénario dramatique  

**Les deux** te montrent le système 3-agents en action !

---

## 🎉 CHOISIS TON AVENTURE

```bash
# Exploration tranquille
./demo.sh

# Challenge immédiat
./arena_challenge.sh
```

**Amuse-toi bien ! 🚀**
