# Mon Premier Agent A2A

## Ce que vous allez faire

Lancer votre premier agent intelligent en **2 minutes** et lui poser votre première question.

**Aucune compétence technique requise** - juste copier/coller des commandes.

---

## Étape 1 : Ouvrir le terminal

**Sur Mac** : 
- Appuyez sur `Cmd + Espace`
- Tapez "Terminal"
- Appuyez sur `Entrée`

**Sur Windows** :
- Appuyez sur `Windows + R`
- Tapez "cmd"
- Appuyez sur `Entrée`

---

## Étape 2 : Aller dans le bon dossier

Copiez cette ligne et collez-la dans le terminal :

```bash
cd ~/stageocto/ucp-merchant-test
```

Appuyez sur `Entrée`.

---

## Étape 3 : Lancer l'agent

Copiez et collez cette commande :

```bash
./scripts/start-agents.sh
```

Appuyez sur `Entrée`.

**Vous devriez voir** :
```
✅ Customer Growth Agent started (port 9001)
✅ Competitiveness Agent started (port 9002)
✅ Dashboard started (port 8080)
```

---

## Étape 4 : Ouvrir l'interface

**Dans votre navigateur** (Chrome, Safari, Firefox...), allez sur :

```
http://localhost:8080
```

**Vous verrez** une page avec deux agents disponibles.

---

## Étape 5 : Poser votre première question

### Test 1 : Analyser un client

1. Cliquez sur **"Customer Growth Agent"**
2. Dans la zone de texte, copiez :
   ```
   elsi
   ```
3. Cliquez sur **"Analyser"**

**Résultat** : L'agent vous dit que "elsi" est un client Gold ayant dépensé $850.

### Test 2 : Vérifier un prix

1. Cliquez sur **"Competitiveness Agent"**
2. Dans la zone de texte, copiez :
   ```
   laptop
   ```
3. Entrez un prix : `1000`
4. Cliquez sur **"Analyser"**

**Résultat** : L'agent vous dit si votre prix est compétitif par rapport au marché.

---

## Étape 6 : Arrêter les agents

Quand vous avez fini :

```bash
./scripts/stop-agents.sh
```

---

## 🎉 Félicitations !

Vous venez de :
- ✅ Lancer 2 agents intelligents
- ✅ Leur poser des questions
- ✅ Obtenir des recommandations business

**Prochaine étape** : [Comment tester d'autres clients](../how-to/tester-clients.md)
