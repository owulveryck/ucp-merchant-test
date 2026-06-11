# ADR-0012: Mock Data Sources pour Mode Standalone

**Date**: 2026-06-09  
**Statut**: ✅ Accepté  
**Décideurs**: Équipe Technique OCTO  
**Tags**: `testing`, `mock`, `standalone`, `demo`

## Contexte

Les agents A2A indépendants (ADR-0011) doivent fonctionner **sans base de données** pour :
- Démos clients sans setup infrastructure
- Tests reproductibles
- Développement local sans dépendances

**Problème** : Comment fournir des données réalistes aux agents sans connexion BDD ?

## Décision

Implémenter des **Mock Data Sources** intégrés dans le code avec données de test prédéfinies.

### Implémentation

#### 1. MockCustomerDataSource

**Fichier** : `pkg/pricing-unified/datasources/mock_customer_data.go`

**Clients de test** :
```go
"elsi": {
    CustomerID:       "elsi",
    TotalSpent:       85000,   // $850 - Gold tier
    PurchaseCount:    8,
    LastPurchaseDays: 10,
}
"olwu": {
    CustomerID:       "olwu",
    TotalSpent:       120000,  // $1200 - Premium tier
    PurchaseCount:    15,
    LastPurchaseDays: 7,
}
"lja": {
    CustomerID:       "lja",
    TotalSpent:       5000,    // $50 - Standard tier
    PurchaseCount:    1,
    LastPurchaseDays: 120,
}
"manu": {
    CustomerID:       "manu",
    TotalSpent:       35000,   // $350 - Silver tier
    PurchaseCount:    4,
    LastPurchaseDays: 20,
}
```

**Tiers calculés automatiquement** :
- Standard : < $100
- Silver : $100-$499
- Gold : $500-$999
- Premium : ≥ $1000

#### 2. MockCompetitorPriceSource

**Fichier** : `cmd/competitiveness-agent/mock_price_source.go`

**Produits avec concurrents** :
```go
"laptop": {
    {MerchantID: "competitor_a", MerchantName: "TechStore", Price: 95000},   // $950
    {MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 105000},    // $1050
    {MerchantID: "competitor_c", MerchantName: "Amazon", Price: 98000},      // $980
}
"mouse": {
    {MerchantID: "competitor_a", MerchantName: "TechStore", Price: 2500},    // $25
    {MerchantID: "competitor_b", MerchantName: "BestBuy", Price: 3000},      // $30
}
"keyboard": {...}
"monitor": {...}
```

#### 3. MockHistoryStore

**Implémentation** : No-op (pas d'historique en mode standalone)
- `GetPriceHistory()` → retourne liste vide
- `GetTrend()` → retourne tendance "stable"
- `RecordPrice()` → no-op

**Raison** : L'historique nécessiterait un stockage persistant, contradictoire avec le mode standalone.

### Interface unifiée

**Pattern Strategy** :
```go
// Interface (production + mock)
type CustomerDataSource interface {
    GetCustomerProfile(customerID string) (CustomerProfile, error)
}

// Utilisation dans l'agent
dataSource := datasources.NewMockCustomerDataSource()  // Standalone
// OU
dataSource := datasources.NewPostgresDataSource(db)    // Production

agent := agents.NewCustomerGrowthAgent(dataSource)
```

## Conséquences

### Positives

**Simplicité**
- ✅ Zéro configuration (pas de BDD, fichiers config, env vars)
- ✅ Démarrage instantané (`./bin/customer-growth-agent`)

**Reproductibilité**
- ✅ Tests toujours identiques (données figées)
- ✅ Démos prévisibles (pas de surprises client)

**Développement**
- ✅ Onboarding rapide (git clone → go run)
- ✅ Pas de docker-compose pour développer

**Documentation**
- ✅ Exemples vivants dans le code (clients elsi, olwu, lja, manu = tutoriel)

### Négatives

**Limitations**
- ❌ Pas de CRUD (données read-only hardcodées)
- ❌ Jeu de données limité (4 clients, 4 produits)
- ❌ Pas d'historique réel

**Drift avec production**
- ⚠️ Risque : mock data diverge de la vraie structure BDD
- ⚠️ Mitigation : Tests d'intégration avec vraie BDD en CI/CD

**Maintenance**
- ❌ Ajouter un client = modifier le code + recompiler
- ⚠️ Mitigation : Futurs ADR pour mock data YAML/JSON externalisé

## Alternatives considérées

### 1. SQLite embarqué

**Pour** : Vraie BDD SQL, CRUD complet  
**Contre** : Fichier à gérer, schéma à migrer, complexité setup  
**Rejet** : Trop complexe pour des démos rapides

### 2. Fichiers JSON/YAML

**Pour** : Éditable sans recompilation  
**Contre** : Parsing errors possibles, chemins relatifs fragiles  
**Rejet** : Le gain (édition sans recompile) ne justifie pas la fragilité

### 3. API mock externe (mockapi.io, wiremock)

**Pour** : Réalisme accru (vraies requêtes HTTP)  
**Contre** : Dépendance réseau, latence, service externe  
**Rejet** : Contradictoire avec "standalone" (pas de dépendance)

### 4. Embedded etcd/bolt

**Pour** : Stockage persistant key-value  
**Contre** : Overhead mémoire, complexité, pas de SQL  
**Rejet** : Over-engineering pour 4 clients de test

## Migration vers production

**Code agent inchangé** (interface identique) :

```go
// Standalone (mock)
func NewCustomerGrowthA2AAgent() *CustomerGrowthA2AAgent {
    dataSource := datasources.NewMockCustomerDataSource()
    agent := agents.NewCustomerGrowthAgent(dataSource, ...)
    return &CustomerGrowthA2AAgent{agent: agent}
}

// Production (BDD)
func NewCustomerGrowthA2AAgent(db *sql.DB) *CustomerGrowthA2AAgent {
    dataSource := datasources.NewPostgresDataSource(db)
    agent := agents.NewCustomerGrowthAgent(dataSource, ...)  // MÊME CODE
    return &CustomerGrowthA2AAgent{agent: agent}
}
```

**Seul le `main.go` change** :
```go
// Standalone
agent := NewCustomerGrowthA2AAgent()

// Production
db := connectDB()
agent := NewCustomerGrowthA2AAgent(db)
```

## Métriques de succès

- ✅ Temps de premier démarrage : < 5s (atteint : ~1s)
- ✅ Taille binaire avec mock : < 15 MB (atteint : ~10 MB)
- ✅ Couverture de test avec mock : > 80% (à mesurer)

## Évolution future

**Phase 1 (actuel)** : Mock data hardcodé  
**Phase 2** : Mock data YAML/JSON externe (optionnel)  
**Phase 3** : Générateur de données aléatoires (Faker)  
**Phase 4** : Mode "recording" (enregistrer vraies réponses BDD → replay)

## Liens

- Code : `pkg/pricing-unified/datasources/mock_customer_data.go`
- Code : `cmd/competitiveness-agent/mock_price_source.go`
- ADR parent : ADR-0011 (Agents A2A Indépendants)
- Pattern : Strategy Pattern pour DataSource
- Commits : `d6fa715` (clients démo), `a891f01` (mock prices)

## Notes

Les mock data sources sont **critiques pour l'adoption** :
- Un prospect peut tester en 30s sans setup
- Les développeurs peuvent contribuer sans BDD locale
- Les démos marchent offline (train, avion, hôtel sans WiFi)

**Principe** : "La simplicité bat la puissance pour les 80% de cas d'usage".
