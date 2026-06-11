# Voir les logs de décision

## Pourquoi voir les logs ?

Les logs vous montrent **comment les agents prennent leurs décisions** en temps réel :

- 🕵️ Quels prix les concurrents proposent
- 📊 Comment l'analyse se déroule
- 🎯 Quelle stratégie est choisie
- ✅ Si la marge est validée
- 🛒 Décision finale : acheter ou pas

---

## Comment voir les logs ?

### Méthode 1 : En temps réel (recommandé)

```bash
tail -f logs/arena.log
```

**Vous verrez** les décisions défiler en direct.

**Pour arrêter** : Appuyez sur `Ctrl + C`

---

### Méthode 2 : Voir tout l'historique

```bash
cat logs/arena.log
```

**Vous verrez** toutes les décisions depuis le démarrage.

---

## Exemple de log

```
[2026-06-11 10:15:32] 🕵️ L'Espion récupère les prix pour 'laptop'
  → Concurrent A: $1000
  → Concurrent B: $1050
  → Concurrent C: $950
  → Prix min: $950, Prix max: $1050

[2026-06-11 10:15:33] 📊 L'Analyste évalue le marché
  → Fourchette: $950 - $1050
  → Position moyenne: $1000
  → Notre proposition: $980
  → Classement: 2/4

[2026-06-11 10:15:34] 🎯 Le Stratège recommande
  → Stratégie: Match Lowest
  → Prix optimal: $950
  → Chance de victoire: 85%

[2026-06-11 10:15:35] ✅ Le Contrôleur valide
  → Coût produit: $800
  → Prix vente: $950
  → Marge: $150 (15.8%)
  → Décision: VALIDÉ ✅

[2026-06-11 10:15:36] 🛒 Shopping Agent décide
  → ACHAT à $950
  → Résultat: GAGNÉ 🏆
```

---

## Comprendre les symboles

| Symbole | Signification |
|---------|---------------|
| 🕵️ | L'Espion collecte les données |
| 📊 | L'Analyste évalue |
| 🎯 | Le Stratège recommande |
| ✅ | Le Contrôleur valide |
| ❌ | Le Contrôleur rejette (marge trop faible) |
| 🛒 | Shopping Agent prend la décision finale |
| 🏆 | Victoire (meilleur prix) |
| 😞 | Défaite (concurrent moins cher) |

---

## Logs spécifiques

### Voir uniquement les victoires

```bash
grep "GAGNÉ" logs/arena.log
```

### Voir uniquement les décisions rejetées

```bash
grep "REJETÉ" logs/arena.log
```

### Voir les recommandations du Stratège

```bash
grep "Stratège recommande" logs/arena.log
```

---

## Logs de chaque composant

L'Arena génère 3 fichiers de logs :

| Fichier | Contenu |
|---------|---------|
| `logs/arena.log` | Décisions des 4 agents |
| `logs/shopping-graph.log` | Orchestration globale |
| `logs/obs-hub.log` | Métriques et observabilité |

**Pour voir tous les logs en même temps** :

```bash
tail -f logs/*.log
```

---

## Que faire en cas d'erreur ?

### Erreur : "No such file or directory"

**Cause** : Les agents ne sont pas lancés

**Solution** :
```bash
./scripts/start-demo.sh
```

Attendez 10 secondes puis relancez :
```bash
tail -f logs/arena.log
```

---

### Erreur : Le fichier est vide

**Cause** : L'Arena n'a pas encore pris de décision

**Solution** : Attendez quelques secondes. L'Arena prend des décisions toutes les 5-10 secondes.

---

## Prochaine étape

[Arrêter l'Arena](arreter-arena.md)
