# How-to : Configurer un Nouveau Marchand

## Objectif

Ajouter un 6ème marchand à l'Arena ou à la démo shopping.

## Prérequis

- Projet compilé
- Connaissance de base JSON/CSV

## Étape 1 : Créer le répertoire de données

```bash
cd ~/stageocto/ucp-merchant-test/demo/data
mkdir merchant_d
cd merchant_d
```

## Étape 2 : Créer le catalog JSON

Créer `catalog.json` :

```json
{
  "products": [
    {
      "id": "laptop_pro",
      "title": "Professional Laptop",
      "price": 129900,
      "currency": "USD",
      "image_url": "https://example.com/laptop.jpg",
      "in_stock": true,
      "quantity": 50
    },
    {
      "id": "mouse_wireless",
      "title": "Wireless Mouse",
      "price": 2999,
      "currency": "USD",
      "image_url": "https://example.com/mouse.jpg",
      "in_stock": true,
      "quantity": 200
    }
  ],
  "discounts": [
    {
      "code": "TECH10",
      "type": "percentage",
      "value": 10,
      "description": "10% off all tech"
    },
    {
      "code": "SAVE50",
      "type": "fixed",
      "value": 5000,
      "description": "$50 off"
    }
  ],
  "customers": [
    {
      "id": "cust_tech_1",
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "tier": "gold"
    }
  ],
  "shipping_rates": [
    {
      "country": "US",
      "service_level": "standard",
      "price": 999,
      "description": "Standard Shipping (5-7 days)"
    },
    {
      "country": "US",
      "service_level": "express",
      "price": 1999,
      "description": "Express Shipping (2-3 days)"
    }
  ],
  "free_shipping_rules": [
    {
      "type": "subtotal_threshold",
      "threshold": 10000,
      "description": "Free shipping over $100"
    }
  ]
}
```

**Note** : Les prix sont en centimes (ex: 129900 = $1,299.00).

## Étape 3 : Lancer le nouveau marchand

```bash
cd ~/stageocto/ucp-merchant-test

demo/bin/merchant \
  --port 8185 \
  --data-dir demo/data/merchant_d \
  --data-format json \
  --merchant-name TechStore \
  --simulation-secret tech-secret
```

## Étape 4 : Vérifier le démarrage

```bash
# Test discovery
curl http://localhost:8185/.well-known/ucp | jq

# Test catalog
curl http://localhost:8185/shopping-api/catalog | jq

# Test checkout
curl -X POST http://localhost:8185/shopping-api/checkout-sessions \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"item_id": "laptop_pro", "quantity": 1}]
  }' | jq
```

## Étape 5 : Intégrer au Shopping Graph

Modifier `demo/config/shopping_graph.yaml` :

```yaml
merchants:
  - name: SuperShop
    url: http://localhost:8182
  - name: MegaMart
    url: http://localhost:8183
  - name: BudgetBuy
    url: http://localhost:8184
  - name: TechStore      # NOUVEAU
    url: http://localhost:8185

polling_interval: 30s
```

Redémarrer le Shopping Graph :

```bash
demo/bin/shopping-graph \
  --port 9000 \
  --config demo/config/shopping_graph.yaml \
  --obs-url http://localhost:9002
```

## Étape 6 : Vérifier l'intégration

```bash
# Rechercher un produit du nouveau marchand
curl "http://localhost:9000/search?q=laptop" | jq

# Résultat attendu : produits de TechStore dans les résultats
```

## Étape 7 : Ajouter à l'Arena (optionnel)

Modifier `demo/cmd/arena/main.go` pour ajouter TechStore aux tenants :

```go
tenants := []struct {
    name string
    port int
    dataDir string
}{
    {"MegaStore", 8182, "demo/data/merchant_a"},
    {"PrixCassés", 8183, "demo/data/merchant_b"},
    {"SuperDeals", 8184, "demo/data/merchant_c"},
    {"TopPrix", 8185, "demo/data/merchant_d"},  // TechStore
    // ...
}
```

Recompiler l'arena :

```bash
go build -o demo/bin/arena ./demo/cmd/arena
demo/bin/arena
```

## Configuration avancée

### Pricing Agent personnalisé

Créer `demo/data/merchant_d/pricing_config.json` :

```json
{
  "min_margin_percent": 15,
  "default_strategy": "premium",
  "vip_discount_percent": 12,
  "competitive_threshold": 0.95
}
```

### Stratégies par produit

```json
{
  "product_strategies": {
    "laptop_pro": {
      "strategy": "premium",
      "min_margin": 20
    },
    "mouse_wireless": {
      "strategy": "aggressive",
      "min_margin": 5
    }
  }
}
```

## Format CSV (alternative)

Si vous préférez CSV au JSON, créez ces fichiers dans `merchant_d/` :

**products.csv** :
```csv
id,title,price,currency,image_url,in_stock,quantity
laptop_pro,Professional Laptop,129900,USD,https://example.com/laptop.jpg,true,50
mouse_wireless,Wireless Mouse,2999,USD,https://example.com/mouse.jpg,true,200
```

**discounts.csv** :
```csv
code,type,value,description
TECH10,percentage,10,10% off all tech
SAVE50,fixed,5000,$50 off
```

Lancer avec :
```bash
demo/bin/merchant \
  --port 8185 \
  --data-dir demo/data/merchant_d \
  --data-format csv \
  --merchant-name TechStore
```

## Troubleshooting

**Le marchand ne démarre pas** :
- Port déjà utilisé ? Vérifiez : `lsof -i :8185`
- Fichier catalog.json invalide ? Validez JSON : `jq . catalog.json`

**Produits introuvables dans Shopping Graph** :
- Le Shopping Graph poll toutes les 30s, attendez
- Vérifiez les logs : `tail -f logs/shopping-graph.log`
- Forcez un refresh : `curl -X POST http://localhost:9000/admin/refresh`

**Erreurs de pricing** :
- Vérifiez que `price` est en centimes, pas en dollars
- Vérifiez `min_margin_percent` dans la config

## Résumé

✅ Créer `demo/data/merchant_X/catalog.json`  
✅ Lancer avec `--port`, `--data-dir`, `--merchant-name`  
✅ Ajouter à `shopping_graph.yaml`  
✅ Redémarrer Shopping Graph  
✅ Vérifier via `/search`  

Votre nouveau marchand est maintenant intégré à l'écosystème !
