# Tester différents clients

## Clients disponibles

Vous avez **4 clients de test** prêts à utiliser :

| Nom | Profil | Budget dépensé |
|-----|--------|----------------|
| **elsi** | Client fidèle Gold | $850 |
| **olwu** | VIP Premium | $1200 |
| **lja** | Nouveau client | $50 |
| **manu** | Client régulier Silver | $350 |

---

## Comment tester un client

### Via le Dashboard (facile)

1. Ouvrez http://localhost:8080
2. Cliquez sur **"Customer Growth Agent"**
3. Entrez le nom du client : `olwu`
4. Cliquez sur **"Analyser"**

### Via le terminal (avancé)

```bash
curl -X POST http://localhost:9001/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "analyze_customer",
    "params": {"customer_id": "olwu"},
    "id": 1
  }'
```

---

## Ce que vous apprenez pour chaque client

L'agent vous donne :

- ✅ **Tier** : Standard, Silver, Gold ou Premium
- ✅ **Recommandation** : Faut-il garder ce client ?
- ✅ **Réduction suggérée** : Quel % de remise offrir ?
- ✅ **Valeur totale** : Combien le client a dépensé

---

## Exemples de résultats

### Client VIP (olwu)
```
✅ Client Premium
✅ À conserver absolument
💰 Réduction : 15%
📊 Valeur : $1200
```

### Nouveau client (lja)
```
⚠️ Client Standard
❌ Pas prioritaire
💰 Réduction : 0%
📊 Valeur : $50
```

---

## Prochaine étape

[Comment tester les prix des produits](tester-prix.md)
