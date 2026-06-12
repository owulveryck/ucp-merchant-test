# Tester les prix des produits

## Produits disponibles

Vous avez **4 produits** avec des concurrents :

| Produit | Prix du marché |
|---------|----------------|
| **laptop** | $950 - $1050 |
| **mouse** | $25 - $30 |
| **keyboard** | $68 - $75 |
| **monitor** | $350 - $380 |

---

## Comment vérifier si un prix est compétitif

### Via le Dashboard (facile)

1. Ouvrez http://localhost:8080
2. Cliquez sur **"Competitiveness Agent"**
3. Choisissez un produit : `laptop`
4. Entrez votre prix : `1000`
5. Cliquez sur **"Analyser"**

### Via le terminal (avancé)

```bash
curl -X POST http://localhost:9002/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_competitiveness",
    "params": {
      "product_id": "laptop",
      "price": 100000
    },
    "id": 1
  }'
```

**Note** : Le prix est en centimes (100000 = $1000)

---

## Ce que l'agent vous dit

- ✅ **Est-ce compétitif ?** : Oui/Non
- 📊 **Position marché** : Classement vs concurrents
- 💡 **Stratégie recommandée** : Match, Under, Premium
- 💰 **Prix optimal** : Quel prix fixer ?
- 📈 **Marge** : Votre profit

---

## Exemples de scénarios

### Scénario 1 : Prix trop élevé

**Vous testez** : Laptop à $1100

**L'agent répond** :
```
❌ Prix NON compétitif
📊 Position : 4/4 (dernier)
💡 Stratégie : Baisser à $1000
💰 Prix optimal : $1000
```

### Scénario 2 : Prix compétitif

**Vous testez** : Laptop à $990

**L'agent répond** :
```
✅ Prix compétitif
📊 Position : 1/4 (meilleur)
💡 Stratégie : Match lowest
💰 Prix optimal : $990
📈 Marge : 15%
```

### Scénario 3 : Prix ultra-bas

**Vous testez** : Laptop à $800

**L'agent répond** :
```
⚠️ Prix trop bas
📊 Position : 1/4 mais marge négative
💡 Stratégie : Remonter à $950
💰 Prix optimal : $950
```

---

## Astuce : Combiner les deux agents

1. **Étape 1** : Analyser le client (`olwu` = Premium)
   → Résultat : Client VIP, réduction 15%

2. **Étape 2** : Vérifier le prix (`laptop` à $1000)
   → Résultat : Prix compétitif

3. **Décision finale** : 
   - Prix de base : $1000
   - Réduction VIP : -15% = $850
   - Prix final : **$850** pour fidéliser un client Premium

---

## Prochaine étape

[Arrêter les agents](arreter-agents.md)
