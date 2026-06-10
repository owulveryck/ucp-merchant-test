# Système Multi-Agents Unifié - Vue d'ensemble

## 🎯 Ce que c'est

Un système de tarification dynamique basé sur 3 agents intelligents qui collaborent pour décider du meilleur prix à offrir à un client.

## 🚀 Démarrage ultra-rapide

```bash
# Test en 2 secondes (sans interface)
./test_quick.sh

# OU démo complète (avec interface web)
./run_unified_demo.sh
# → Ouvre http://localhost:8888
```

## 🤖 Les 3 Agents

### AGENT 1 : VENDEUR (Orchestrateur)
**Rôle** : "Quel prix donner à ce client pour cet item ?"
- Coordonne les agents 2 et 3
- Synthétise leurs décisions
- Décide du prix final

### AGENT 2 : CUSTOMER GROWTH
**Rôle** : "Est-ce un client à garder ?"
- Analyse l'historique client
- Calcule la valeur vie (LTV)
- Décide : OUI/NON + % de réduction

**Tiers** :
- 🌟 PREMIUM (≥$1000) : 15% de réduction
- 🥇 GOLD ($500-999) : 10% de réduction  
- 🥈 SILVER ($200-499) : 5% de réduction
- ⚪ STANDARD (<$200) : 0% de réduction

### AGENT 3 : COMPÉTITIVITÉ
**Rôle** : "Sommes-nous compétitifs sur cet item ?"
- Analyse les prix concurrents
- Détermine notre position marché
- Recommande un prix compétitif

**Architecture hybride** : Enveloppe le système 4-agents existant :
- Agent Intel : Prix concurrents
- Agent Insight : Analyse marché
- Agent Strategy : Recommandation
- Agent Validation : Marges

## 📊 Flux de décision

```
1. Client demande un prix
   ↓
2. AGENT 1 consulte AGENT 2
   "Garder ce client ?"
   ↓
3. AGENT 2 répond
   "OUI, tier PREMIUM, -15%"
   ↓
4. AGENT 1 consulte AGENT 3
   "Prix compétitif ?"
   ↓
5. AGENT 3 répond
   "Prix marché : $57"
   ↓
6. AGENT 1 synthétise
   "$57 - 15% = $48.45"
   ↓
7. Client reçoit $48.45
```

## 🧪 Exemple concret

**Scénario** : Client premium demande un casque à $60

```
👤 AGENT 2: CUSTOMER GROWTH
   ✅ Garder ce client : OUI
   Tier : premium
   Réduction suggérée : 15%
   Lifetime Value : $1,500

📊 AGENT 3: COMPÉTITIVITÉ
   ✅ Compétitif : OUI
   Position marché : 2/5
   Prix concurrent le plus bas : $55
   Prix recommandé : $57

🎯 AGENT 1: VENDEUR (DÉCISION)
   Prix de base : $60.00
   Prix compétitif : $57.00
   Bonus VIP : -15%
   ────────────────────────
   PRIX FINAL : $48.45
   Marge : 3%
   Stratégie : vip_retention
```

## 📁 Structure du code

```
pkg/pricing-unified/
├── orchestrator.go              # Agent 1 (Vendeur)
├── models/types.go              # Types de données
├── agents/
│   ├── customer_growth.go       # Agent 2
│   └── competitiveness.go       # Agent 3 (wrapper 4-agents)
└── example/main.go              # Démo standalone

Scripts de lancement :
├── test_quick.sh                # Test rapide
├── run_unified_demo.sh          # Démo complète
├── DEMARRAGE.md                 # Guide de démarrage
└── TEST_VALUES.md               # Valeurs de test détaillées
```

## 🎮 Tester sur l'interface

1. Lance la démo : `./run_unified_demo.sh`
2. Ouvre : http://localhost:8888
3. Clique sur **"Test AUTO_COMPETE"**
4. Utilise ces Customer IDs :
   - `premium_vip_001` → Réduction 15%
   - `gold_customer_002` → Réduction 10%
   - `silver_customer_003` → Réduction 5%
   - `standard_customer_999` → Pas de réduction

**Code promo magique** : `AUTO_COMPETE`

## ✅ Avantages de cette architecture

✅ **Préserve l'existant** : Le système 4-agents continue de fonctionner  
✅ **Simplifie l'interface** : 3 agents au lieu de 4  
✅ **Ajoute la rétention** : Nouveau layer Customer Growth  
✅ **Mode dégradé** : Fonctionne même sans Shopping Graph  
✅ **Testable** : Chaque agent peut être testé indépendamment  

## 📚 Documentation

- **DEMARRAGE.md** : Guide de démarrage rapide
- **TEST_VALUES.md** : Tous les scénarios de test
- **pkg/pricing-unified/README.md** : Documentation technique complète

## 🛠️ Pour les développeurs

### Utiliser le système dans ton code

```go
import (
    pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified"
    "github.com/owulveryck/ucp-merchant-test/pkg/pricing-unified/models"
)

// Créer l'orchestrateur vendeur
vendeur := pricing.NewVendorOrchestrator(agent2, agent3)

// Demander un prix
request := models.PricingRequest{
    ProductID:  "casque_audio",
    CustomerID: "premium_vip_001",
    BasePrice:  6000,  // $60 en centimes
    CostPrice:  5000,  // $50 en centimes
}

decision, err := vendeur.DeterminePricing(request)

// decision.FinalPrice = prix final en centimes
// decision.CustomerGrowth = décision de l'Agent 2
// decision.Competitiveness = décision de l'Agent 3
// decision.DecisionReasoning = raisonnement complet
```

### Implémenter ton propre CustomerDataSource

```go
type MyCustomerData struct {
    db *sql.DB
}

func (m *MyCustomerData) GetCustomerProfile(customerID string) (agents.CustomerProfile, error) {
    // Récupère depuis ta base de données
    return agents.CustomerProfile{
        CustomerID:       customerID,
        TotalSpent:       totalSpent,      // en centimes
        PurchaseCount:    purchaseCount,
        LastPurchaseDays: lastPurchaseDays,
    }, nil
}
```

## 🔧 Troubleshooting

**Port déjà utilisé** :
```bash
lsof -ti:9000 | xargs kill -9  # Shopping Graph
lsof -ti:8888 | xargs kill -9  # Arena
```

**Shopping Graph connection refused** :
Normal si tu utilises `test_quick.sh`. Le système fonctionne en mode dégradé (pas de données concurrents).

**Dashboard ne charge pas** :
```bash
cd demo
go build -o bin/arena ./cmd/arena
./bin/arena --port 8888
```

## 📞 Support

Questions ? Problèmes ?
→ Consulte **TEST_VALUES.md** pour les scénarios détaillés
→ Consulte **pkg/pricing-unified/README.md** pour la doc technique

---

**TL;DR** : `./test_quick.sh` pour tester rapidement, `./run_unified_demo.sh` pour la démo complète avec UI.
