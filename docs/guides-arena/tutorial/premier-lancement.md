# Mon Premier Lancement Arena

## Ce que vous allez faire

Lancer l'environnement Arena complet avec **4 agents intelligents** qui travaillent ensemble pour gagner des compétitions d'achat.

**Temps estimé** : 5 minutes

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

## Étape 3 : Lancer l'Arena

Copiez et collez cette commande :

```bash
./scripts/start-demo.sh
```

Appuyez sur `Entrée`.

**Vous devriez voir** :
```
✅ Shopping Graph started
✅ Observability Hub started  
✅ Arena started
🎮 Arena is ready!
```

---

## Étape 4 : Voir l'Arena en action

L'Arena fonctionne automatiquement. Les agents analysent les produits et prennent des décisions d'achat intelligentes.

**Consulter les logs** :
```bash
tail -f logs/arena.log
```

**Vous verrez** :
- 🕵️ L'Espion : Récupère les prix des concurrents
- 📊 L'Analyste : Analyse le marché
- 🎯 Le Stratège : Recommande une stratégie de prix
- ✅ Le Contrôleur : Valide que la marge est acceptable
- 🛒 Shopping Agent : Décide d'acheter ou non

Appuyez sur `Ctrl + C` pour arrêter de voir les logs.

---

## Étape 5 : Comprendre ce qui se passe

L'Arena simule un **environnement de compétition** :

```
┌─────────────────────────────────────────┐
│         ENVIRONNEMENT ARENA             │
│                                         │
│  🏪 Marchands (vous + concurrents)     │
│  📦 Produits disponibles                │
│  💰 Prix qui changent                   │
│                                         │
│  🎯 Objectif : Acheter au meilleur prix│
└─────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────┐
│       VOS 4 AGENTS INTELLIGENTS         │
│                                         │
│  🕵️ L'Espion → Prix concurrents        │
│  📊 L'Analyste → Position marché        │
│  🎯 Le Stratège → Stratégie prix        │
│  ✅ Le Contrôleur → Valide marge        │
│                                         │
│  Résultat : Décision d'achat optimale  │
└─────────────────────────────────────────┘
```

---

## Étape 6 : Arrêter l'Arena

Quand vous avez fini :

```bash
./scripts/stop-demo.sh
```

**Vous verrez** :
```
✅ Arena stopped
✅ Observability Hub stopped
✅ Shopping Graph stopped
```

---

## 🎉 Félicitations !

Vous venez de :
- ✅ Lancer l'environnement Arena complet
- ✅ Voir 4 agents travailler ensemble
- ✅ Comprendre le flux de décision

**Prochaine étape** : [Comprendre les 4 agents](../explanation/les-4-agents.md)
