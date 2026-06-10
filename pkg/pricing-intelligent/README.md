# Système Multi-Agents de Pricing Intelligent

**Le bon prix, au bon client, selon le marché**

## Architecture

```
Acheteur
    ↓
Agent Vendeur (Orchestrateur)
    ├── Agent Fidélité (analyse client)
    └── Agent Compétitivité (analyse marché)
```

## Agents

### 🎯 Agent Vendeur (Orchestrateur)
- **Intention** : Quel prix optimiser pour maximiser profit et vente ?
- **Décision** : Prix final calculé selon valeur client + position marché

### 💎 Agent Fidélité
- **Intention** : Ce client mérite-t-il un prix préférentiel ?
- **Analyse** : Historique d'achats, montant dépensé, potentiel
- **Décision** : Client VIP (prix ajusté) ou Standard (prix normal)

### 📊 Agent Compétitivité
- **Intention** : Sommes-nous compétitifs sur ce produit ?
- **Analyse** : Prix concurrents, codes promo, position marché
- **Décision** : Prix pour gagner + Position

## Structure

```
pkg/pricing-intelligent/
├── README.md
├── orchestrator.go          # Agent Vendeur (chef)
├── models/
│   ├── types.go            # Types de données
│   └── interfaces.go       # Interfaces des agents
├── agents/
│   ├── loyalty.go          # Agent Fidélité
│   └── competitiveness.go  # Agent Compétitivité
└── api/
    └── handler.go          # API REST
```

## Utilisation

```go
// Créer l'orchestrateur
orchestrator := NewOrchestrator(loyaltyAgent, competitivenessAgent)

// Calculer prix optimal
result := orchestrator.CalculateOptimalPrice(customerID, productID, basePrice)

// Résultat
// result.FinalPrice
// result.Reasoning
// result.IsVIP
```
