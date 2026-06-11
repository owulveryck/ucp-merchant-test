# Référence : Clients de test

## Liste complète

| ID | Nom complet | Budget total | Nombre achats | Dernière activité | Tier |
|----|-------------|--------------|---------------|-------------------|------|
| `elsi` | Elsi | $850 | 8 achats | 10 jours | **Gold** |
| `olwu` | Olwu | $1200 | 15 achats | 7 jours | **Premium** |
| `lja` | Lja | $50 | 1 achat | 120 jours | **Standard** |
| `manu` | Manu | $350 | 4 achats | 20 jours | **Silver** |

---

## Tiers expliqués

### Standard (< $100)
- Nouveau client ou inactif
- Réduction : 0%
- Priorité : Basse

### Silver ($100 - $499)
- Client régulier
- Réduction : 5%
- Priorité : Moyenne

### Gold ($500 - $999)
- Client fidèle
- Réduction : 10%
- Priorité : Haute

### Premium (≥ $1000)
- Client VIP
- Réduction : 15%
- Priorité : Maximale

---

## Pour tester

Utilisez l'ID dans vos requêtes :

**Dashboard** : Entrez `elsi`, `olwu`, `lja` ou `manu`

**Terminal** :
```bash
curl -X POST http://localhost:9001/a2a \
  -d '{"jsonrpc":"2.0","method":"analyze_customer","params":{"customer_id":"elsi"},"id":1}'
```

---

## Retour

[Comment tester les clients](../how-to/tester-clients.md)
