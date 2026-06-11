# Référence : Produits de test

## Liste complète

| ID | Nom | Prix marché minimum | Prix marché maximum | Nombre concurrents |
|----|-----|---------------------|---------------------|-------------------|
| `laptop` | Ordinateur portable | $950 | $1050 | 3 concurrents |
| `mouse` | Souris | $25 | $30 | 2 concurrents |
| `keyboard` | Clavier | $68 | $75 | 3 concurrents |
| `monitor` | Écran | $350 | $380 | 2 concurrents |

---

## Détails par produit

### Laptop (Ordinateur portable)

**Concurrents** :
- Concurrent A : $1000
- Concurrent B : $1050
- Concurrent C : $950

**Prix optimal recommandé** : $990 - $1000

---

### Mouse (Souris)

**Concurrents** :
- Concurrent A : $30
- Concurrent B : $25

**Prix optimal recommandé** : $27 - $28

---

### Keyboard (Clavier)

**Concurrents** :
- Concurrent A : $75
- Concurrent B : $70
- Concurrent C : $68

**Prix optimal recommandé** : $70 - $72

---

### Monitor (Écran)

**Concurrents** :
- Concurrent A : $380
- Concurrent B : $350

**Prix optimal recommandé** : $360 - $370

---

## Format des prix

**⚠️ Important** : Les prix sont en **centimes** dans l'API

| Prix affiché | Prix API |
|--------------|----------|
| $10 | 1000 |
| $100 | 10000 |
| $1000 | 100000 |

**Exemple** :
- Pour tester un laptop à $1000, entrez `100000`
- Pour tester une souris à $28, entrez `2800`

---

## Pour tester

**Dashboard** : Les prix sont automatiquement convertis (entrez `1000` pour $1000)

**Terminal** :
```bash
curl -X POST http://localhost:9002/a2a \
  -d '{
    "jsonrpc":"2.0",
    "method":"analyze_competitiveness",
    "params":{"product_id":"laptop","price":100000},
    "id":1
  }'
```

---

## Retour

[Comment tester les prix](../how-to/tester-prix.md)
