# 🎯 Intelligence Compétitive - Guide d'utilisation

## 🚀 Lancement rapide

Une seule commande pour tout lancer :

```bash
cd /Users/e.g.singer/stageocto/ucp-merchant-test
./run_competitive_demo.sh
```

## 📊 Ce qui a été ajouté

### 1. **Section Intelligence Compétitive dans le Dashboard**

Chaque marchand voit maintenant :
- ✅ **Les prix des concurrents** en temps réel
- ✅ **Une recommandation de prix** pour battre la concurrence
- ✅ **Un bouton "Appliquer ce prix"** pour ajuster instantanément
- ✅ **Le calcul de marge** après ajustement

### 2. **API d'Intelligence Compétitive**

Endpoint : `GET /api/competitive-intel`

Retourne :
```json
{
  "our_price": 7000,
  "our_price_display": "$70.00",
  "lowest_price": 6500,
  "lowest_price_by": "SuperShop",
  "competitors": [
    {
      "merchant_id": "abc123",
      "merchant_name": "SuperShop",
      "price": 6500,
      "price_display": "$65.00",
      "is_us": false
    }
  ],
  "recommended_price": 6175,
  "recommended_price_display": "$61.75",
  "margin_percent": 15,
  "would_win": true,
  "message": "💡 Lower to $61.75 to beat SuperShop and win sales!"
}
```

## 🎮 Comment utiliser

### Étape 1 : Lancer la démo

```bash
./run_competitive_demo.sh
```

### Étape 2 : Ouvrir le navigateur

```
http://localhost:8888/
```

### Étape 3 : Créer des marchands

- Cliquez sur "Register" ou utilisez le formulaire
- Créez 2-3 marchands avec des noms différents
- Exemple : "SuperShop", "MegaMart", "BudgetBuy"

### Étape 4 : Configurer des prix différents

Pour chaque marchand :
1. Allez dans son dashboard (cliquez sur le nom)
2. Ajustez le **slider "Prix"** à une valeur différente
   - Marchand 1 : $70
   - Marchand 2 : $65
   - Marchand 3 : $80

### Étape 5 : Voir l'intelligence compétitive

Dans le dashboard de **Marchand 1** ($70), vous verrez :

```
🎯 Intelligence Compétitive
──────────────────────────
💡 Lower to $61.75 to beat SuperShop and win sales!

Prix suggéré: $61.75
Économie: -$8.25 (marge: 15%)

[✨ Appliquer ce prix]

Concurrents:
SuperShop       $65.00
BudgetBuy       $80.00
Marchand 1      $70.00 [VOUS]
```

### Étape 6 : Appliquer le prix recommandé

Cliquez sur **"✨ Appliquer ce prix"** et le slider s'ajustera automatiquement !

## 🏆 Objectif

**Avoir toujours le meilleur prix** pour que le client agent vous choisisse en priorité.

## 🔄 Rafraîchissement automatique

L'intelligence compétitive se rafraîchit automatiquement **toutes les 10 secondes**.

Si un concurrent baisse son prix, vous verrez immédiatement une nouvelle recommandation.

## 💡 Stratégie de pricing

Le système recommande de :
1. **Battre le concurrent le plus bas de 5%**
2. **Maintenir une marge minimale de 10%**
3. **Ne jamais descendre sous le prix de coût + 10%**

Exemple :
- Concurrent le plus bas : $65.00
- Battre de 5% : $65.00 - $3.25 = $61.75
- Marge : (Prix - Coût) / Prix = 15% ✅

## 🛑 Arrêter la démo

Dans le terminal : **Ctrl+C**

## 📁 Fichiers modifiés

- `demo/cmd/arena/competitive_intel.go` - API d'intelligence compétitive
- `demo/cmd/arena/dashboard.go` - Interface utilisateur
- `demo/cmd/arena/tenant.go` - Endpoint API
- `run_competitive_demo.sh` - Script de lancement

## 🎯 Différence avec l'implémentation précédente

### Avant (AUTO_COMPETE)
- ✅ Automatique via code promo
- ❌ Pas de visibilité sur les concurrents
- ❌ Pas de contrôle manuel

### Maintenant (Intelligence Compétitive)
- ✅ **Visibilité totale** sur les prix concurrents
- ✅ **Recommandations explicites** avec calcul de marge
- ✅ **Contrôle manuel** avec un clic
- ✅ **Rafraîchissement en temps réel**
- ✅ **Interface claire et intuitive**

## 🔧 Debugging

Si la section Intelligence Compétitive affiche "Chargement..." :

1. Vérifiez que le Shopping Graph est lancé :
   ```bash
   curl http://localhost:9000/health
   ```

2. Vérifiez les logs de l'Arena dans le terminal

3. Attendez 15 secondes après la création des marchands (indexation)

---

**Fait avec ❤️ par Claude Code**
