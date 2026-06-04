# Système Multi-Agents Unifié - Architecture Hybride

Ce système implémente une architecture multi-agents pour la tarification dynamique, en préservant le système 4-agents existant tout en ajoutant une nouvelle couche d'orchestration simplifiée.

## Architecture

### AGENT 1: VENDEUR (Orchestrateur Principal)
**Fichier**: `orchestrator.go`

**INTENTION**: "Je suis acheteur lambda et je veux tel item - quel prix lui donner ?"

**FLUX**:
1. Acheteur Lambda → Agent 1 (demande de prix)
2. Agent 1 → Agent 2 (vérifier si client à garder)
3. Agent 1 → Agent 3 (vérifier compétitivité)
4. Agent 2 → Agent 1 (réponse OUI/NON)
5. Agent 3 → Agent 1 (prix compétitif)
6. Agent 1 → Acheteur Lambda (prix final)

**DÉCISION**: Prix final basé sur la synthèse des deux agents

### AGENT 2: CUSTOMER GROWTH
**Fichier**: `agents/customer_growth.go`

**INTENTION**: "Est-ce que lambda est un client que je peux garder ?"

**DONNÉES**:
- Historique achats
- Valeur vie client (LTV)
- Fréquence d'achats
- Recency

**DÉCISION**: OUI/NON + niveau de réduction

**TIERS**:
- **Premium** (≥$1000 dépensés): 15% de réduction
- **Gold** (≥$500 dépensés): 10% de réduction
- **Silver** (≥$200 dépensés): 5% de réduction
- **Standard** (<$200 dépensés): 0% de réduction

### AGENT 3: COMPÉTITIVITÉ
**Fichier**: `agents/competitiveness.go`

**INTENTION**: "Est-ce que je suis compétitif sur cet item ?"

**ARCHITECTURE HYBRIDE**: Cet agent *enveloppe* le système 4-agents existant:
- Agent 1 Intelligence: Prix concurrents
- Agent 2 Insight: Analyse de marché
- Agent 3 Strategy: Recommandation stratégique
- Agent 4 Validation: Validation des marges

**DÉCISION**: Prix compétitif + stratégie à adopter

## Structure des fichiers

```
pkg/pricing-unified/
├── README.md                    # Ce fichier
├── models/
│   └── types.go                 # Types de données (PricingRequest, Decisions, etc.)
├── agents/
│   ├── customer_growth.go       # Agent 2
│   └── competitiveness.go       # Agent 3 (wrapper du système 4-agents)
├── orchestrator.go              # Agent 1 (Vendeur)
└── example/
    └── main.go                  # Démonstration
```

## Utilisation

### Exemple basique

```bash
cd /Users/e.g.singer/stageocto/ucp-merchant-test
go run ./pkg/pricing-unified/example/main.go
```

### Intégration dans votre code

```go
import (
    "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive"
    compAgents "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/agents"
    compModels "github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
    pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified"
    "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/agents"
    "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// 1. Créer le système 4-agents existant
sgClient := competitive.NewShoppingGraphClient("http://localhost:9000")
priceIntel := compAgents.NewPriceIntelligenceAgent(sgClient, "your_merchant_id")
// ... autres agents

orchestrator4Agents := competitive.NewOrchestrator(...)

// 2. Créer les agents unifiés
customerData := &YourCustomerDataSource{}  // Implémente CustomerDataSource
agent2 := agents.NewCustomerGrowthAgent(customerData)

businessConfig := compModels.BusinessConfig{...}
agent3 := agents.NewCompletivenessAgent(orchestrator4Agents, "merchant_id", 5000, businessConfig)

// 3. Créer l'orchestrateur vendeur
vendeur := pricing.NewVendorOrchestrator(agent2, agent3)

// 4. Obtenir un prix
request := models.PricingRequest{
    ProductID:  "product_123",
    CustomerID: "customer_456",
    BasePrice:  6000,  // $60.00
    CostPrice:  5000,  // $50.00
}

decision, err := vendeur.DeterminePricing(request)
if err != nil {
    log.Fatal(err)
}

// decision.FinalPrice contient le prix final
// decision.CustomerGrowth contient la décision de l'Agent 2
// decision.Competitiveness contient la décision de l'Agent 3
// decision.DecisionReasoning contient le raisonnement complet
```

### Implémenter CustomerDataSource

```go
type YourCustomerDataSource struct {
    // vos dépendances (DB, cache, etc.)
}

func (ds *YourCustomerDataSource) GetCustomerProfile(customerID string) (agents.CustomerProfile, error) {
    // Récupérer les données du client depuis votre système
    return agents.CustomerProfile{
        CustomerID:       customerID,
        TotalSpent:       totalSpent,      // en centimes
        PurchaseCount:    purchaseCount,
        LastPurchaseDays: lastPurchaseDays,
    }, nil
}
```

## Exemple de sortie

```
SCÉNARIO: Client Premium demande un produit

👤 AGENT 2: CUSTOMER GROWTH
   Garder ce client ? ✅ OUI
   Tier: premium
   Réduction suggérée: 15%
   Lifetime Value: $1500.00

📊 AGENT 3: COMPÉTITIVITÉ
   Compétitif ? ✅ OUI
   Position: 2/5
   Concurrent le moins cher: $55.00
   Prix recommandé: $57.00

🎯 AGENT 1: VENDEUR (DÉCISION FINALE)
   Prix de base: $60.00
   Prix final offert: $51.00
   Réduction: $9.00 (15%)
   Marge: 1%
   Stratégie: vip_retention

💡 RAISONNEMENT:
   Agent 2 (Customer Growth): OUI, garder ce client - premium
   Agent 3 (Compétitivité): Position 2/5 - Prix recommandé: $57.00
   Bonus fidélité appliqué: -15% (client premium)
   Prix final: $51.00 (-15%)
   Marge: 1%
   Stratégie: vip_retention
```

## Stratégies de prix

L'Agent Vendeur applique 3 stratégies:

1. **vip_retention**: Client de valeur (premium/gold/silver) → bonus fidélité
2. **competitive_pricing**: Position compétitive sur le marché → prix de marché
3. **market_alignment**: Alignement sur le marché sans bonus spécifique

## Avantages de l'architecture hybride

✅ **Préserve le code existant**: Le système 4-agents continue de fonctionner  
✅ **Simplifie l'interface**: 3 agents au lieu de 4 pour l'appelant  
✅ **Ajoute la rétention client**: Nouveau layer de Customer Growth  
✅ **Compatibilité totale**: Aucune modification au système existant requis  
✅ **Testable indépendamment**: Chaque agent peut être testé seul  

## Fallback sans Shopping Graph

Si le Shopping Graph n'est pas disponible, le système fonctionne en mode dégradé:
- Agent 3 utilise le prix de base comme référence
- Agent 2 continue de fonctionner normalement
- Agent 1 fait sa synthèse avec les données disponibles
