# 🎯 Scénarios de Démonstration - Intelligence Compétitive

## 🚀 Comment démarrer

```bash
./run_full_demo.sh
```

Puis allez sur http://localhost:8888

---

## 📋 Scénario 1 : Victoire Simple (RECOMMANDÉ pour démo rapide)

**Objectif :** Prouver que l'outil fonctionne en 2 minutes

### Configuration :

**MarchandA (concurrent)** :
- Prix : $60.00
- Stock : 100
- CPC Bid : $0.50
- Codes promo : `WELCOME10` (type: percentage, value: 10)

**MonMagasin (VOUS)** :
- Prix : $70.00
- Stock : 50
- CPC Bid : $0.50
- Codes promo : aucun

### Étapes :

1. **Créer les marchands** :
   - Cliquez "Rejoindre l'arène" 2 fois
   - Nommez : `MarchandA`, `MonMagasin`

2. **Configurer MarchandA** :
   - Prix : 6000 (= $60)
   - Ajoutez code promo : Code `WELCOME10`, Type `percentage`, Value `10`

3. **Sur MonMagasin** :
   - Prix actuel : 7000 (= $70)
   - Cliquez "💡 Calculer le meilleur prix"

4. **Observer les agents** :
   ```
   Agent 1 (Espion) : "Trouvé 1 concurrent. Le moins cher: $54.00"
   Agent 2 (Analyste) : "⚠️ Vous êtes en position 2/2"
   Agent 3 (Stratège) : "Prix recommandé: $53.00"
   Agent 4 (Contrôleur) : "✅ Validé ! Vous gagnerez 6% de marge"
   ```

5. **Appliquer** :
   - Cliquez "Appliquer ce prix"
   - Prix passe de $70 → $53

6. **Tester avec agent acheteur** :
   ```bash
   cd demo
   go run ./cmd/client --arena-url http://localhost:8888 --query "casque audio"
   ```

### ✅ Résultat attendu :

```
Le prix le plus bas a été trouvé chez MonMagasin pour $53.XX
→ MonMagasin GAGNE ! 🏆
```

---

## 📋 Scénario 2 : Concurrence Agressive (3 marchands)

**Objectif :** Montrer que l'outil bat PLUSIEURS concurrents avec différents codes promo

### Configuration :

**MarchandA** :
- Prix : $60.00
- Code promo : `SAVE15` (percentage, 15) = **$51 effectif**

**MarchandB** :
- Prix : $58.00
- Code promo : `FIXED500` (fixed_amount, 500) = **$53 effectif**

**MonMagasin** :
- Prix : $65.00
- Aucun code promo

### Résultat avec l'outil :

```
Agent 1 : "Le moins cher: $51.00 (MarchandA avec SAVE15)"
Agent 3 : "Prix recommandé: $50.00"
Agent 4 : "✅ Validé avec 0% de marge (prix = coût)"

→ MonMagasin à $50.00
→ Bat MarchandA ($51) et MarchandB ($53)
→ VICTOIRE ! 🏆
```

---

## 📋 Scénario 3 : Protection contre la perte

**Objectif :** Montrer que l'outil NE VEND PAS À PERTE

### Configuration :

**CompetitorX** :
- Prix : $48.00
- Code promo : `MEGA25` (percentage, 25) = **$36 effectif**

**MonMagasin** :
- Prix : $65.00
- Coût : $50.00 (fixé dans le serveur)

### Résultat avec l'outil :

```
Agent 1 : "Le moins cher: $36.00"
Agent 2 : "⚠️ Vous êtes très cher"
Agent 3 : "Il faut baisser à $35.00 pour gagner"
Agent 4 : "❌ REJETÉ: Prix $35 en dessous du coût $50"

→ Aucun changement de prix
→ Vous NE PERDEZ PAS d'argent
→ L'outil vous PROTÈGE ✅
```

**Message à l'utilisateur** :
```
Agent 4 : "❌ Cannot win without selling at loss (target $35 < cost $50)"
```

---

## 📋 Scénario 4 : Déjà le meilleur prix

**Objectif :** Montrer que l'outil ne change rien si vous êtes déjà le moins cher

### Configuration :

**MarchandA** :
- Prix : $70.00

**MonMagasin** :
- Prix : $55.00

### Résultat avec l'outil :

```
Agent 1 : "Trouvé 1 concurrent. Le moins cher: $70.00 (rang 1/2)"
Agent 2 : "✅ Vous êtes le moins cher !"
Agent 3 : "Vous avez déjà le meilleur prix !"
Agent 4 : "✅ Vous êtes déjà le meilleur"

→ Aucun changement nécessaire
→ Vous gagnez DÉJÀ ! 🏆
```

---

## 📋 Scénario 5 : Marge réduite acceptée pour gagner

**Objectif :** Montrer l'arbitrage marge vs victoire

### Configuration :

**MarchandA** :
- Prix : $58.00
- Code promo : `WELCOME10` (10%) = **$52.20 effectif**

**MonMagasin** :
- Prix : $65.00
- Coût : $50.00
- Min marge : 10% (configuré serveur)

### Résultat avec l'outil :

```
Agent 1 : "Le moins cher: $52.20"
Agent 3 : "Prix recommandé: $51.20"
Agent 4 : "⚠️ Marge réduite: 2% (cible: 10%) pour GAGNER"

→ Prix à $51.20
→ Marge seulement 2% au lieu de 10%
→ Mais vous GAGNEZ la vente ! 🏆
```

**Calcul de marge** :
```
Prix final : $51.20
Coût : $50.00
Profit : $1.20
Marge : $1.20 / $51.20 = 2.3%
```

---

## 🎬 Script de Démo Complet (5 minutes)

### Minute 1 : Setup
```bash
./run_full_demo.sh
# Ouvrir http://localhost:8888
```

### Minute 2 : Créer marchands
- Rejoindre 2 fois
- MarchandA : Prix $60, Code WELCOME10 (10%)
- MonMagasin : Prix $70

### Minute 3 : Utiliser l'outil
- Dashboard MonMagasin
- "💡 Calculer le meilleur prix"
- Observer : Prix recommandé $53
- "Appliquer ce prix"

### Minute 4 : Tester
```bash
cd demo
go run ./cmd/client --arena-url http://localhost:8888 --query "casque"
```

### Minute 5 : Montrer résultat
```
MonMagasin: $53.XX ✅ GAGNANT
MarchandA: $60.XX (avec WELCOME10 = $54)

→ Agent a choisi MonMagasin
→ DÉMONSTRATION RÉUSSIE ! 🎉
```

---

## 📊 Tableau Comparatif des Scénarios

| Scénario | Concurrent le moins cher | Prix MonMagasin AVANT | Prix MonMagasin APRÈS | Résultat |
|----------|--------------------------|------------------------|------------------------|----------|
| 1. Simple | $54 (avec promo) | $70 | $53 | ✅ GAGNE |
| 2. Multi-concurrents | $51 (SAVE15) | $65 | $50 | ✅ GAGNE |
| 3. Protection perte | $36 (MEGA25) | $65 | $65 (inchangé) | ✅ PROTÉGÉ |
| 4. Déjà meilleur | $70 | $55 | $55 (inchangé) | ✅ GAGNE |
| 5. Marge réduite | $52.20 (WELCOME10) | $65 | $51.20 | ✅ GAGNE (marge 2%) |

---

## 🔧 Paramètres Serveur

Les paramètres suivants sont définis au lancement dans `run_full_demo.sh` :

```bash
--cost-price 5000        # Coût = $50 (plancher absolu)
--min-margin 10          # Marge cible = 10%
--competitive-pricing    # Active multi-agent
```

**Règles de l'Agent 4 (Contrôleur)** :
- Prix < $50 → ❌ REJETÉ (vente à perte)
- Prix ≥ $50, marge < 10% → ⚠️ ACCEPTÉ pour gagner
- Prix ≥ $50, marge ≥ 10% → ✅ VALIDÉ

---

## 💡 Messages des Agents

### Quand vous GAGNEZ :
```
Agent 1 : "Trouvé X concurrent(s). Le moins cher: $XX"
Agent 2 : "⚠️ Vous êtes en position Y/Z" 
Agent 3 : "Prix recommandé: $XX"
Agent 4 : "✅ Validé ! Vous gagnerez X% de marge"
OU
Agent 4 : "⚠️ Marge réduite: X% (cible: 10%) pour GAGNER"
```

### Quand vous êtes DÉJÀ le meilleur :
```
Agent 1 : "Rang 1/X"
Agent 2 : "✅ Vous êtes le moins cher !"
Agent 3 : "Vous avez déjà le meilleur prix !"
Agent 4 : "✅ Vous êtes déjà le meilleur"
```

### Quand gagner = perte :
```
Agent 3 : "Il faut baisser à $XX pour gagner"
Agent 4 : "❌ Cannot win without selling at loss (target $XX < cost $50)"
```

---

## 🎯 Points Clés à Montrer en Démo

1. ✅ **L'outil détecte les codes promo concurrents** (WELCOME10 → prix effectif $54)
2. ✅ **L'outil recommande un prix gagnant** ($53 < $54)
3. ✅ **L'outil respecte le coût** (ne vend jamais < $50)
4. ✅ **L'outil accepte marge réduite pour GAGNER** (2% au lieu de 10%)
5. ✅ **L'agent acheteur choisit MonMagasin** → VICTOIRE !

---

## 🚨 Troubleshooting

**Prix ne s'affiche pas** :
```bash
# Relancer les services
pkill -f arena
./run_full_demo.sh
```

**Agent acheteur ne trouve rien** :
```bash
# Vérifier que Shopping Graph tourne
curl http://localhost:9000/search -d '{"query":"casque"}' -H "Content-Type: application/json"
```

**Prix ne change pas** :
- Vérifiez que le mode est "Manual" (bouton rouge)
- Cliquez bien "Appliquer ce prix" après calcul
