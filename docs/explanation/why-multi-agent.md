# Explanation : Pourquoi un Système Multi-Agents ?

## Le Problème du Pricing E-commerce

### Approche Traditionnelle (Monolithique)

Un pricing engine classique prend **toutes** les décisions dans un seul système :

```
Input (product, customer, competitors) 
  → Black Box Algorithm
    → Output (price)
```

**Limites** :
- ❌ Logique entremêlée (fidélisation + compétition + marges)
- ❌ Difficile à expliquer ("pourquoi ce prix ?")
- ❌ Peu adaptable (changer une règle impacte tout)
- ❌ Pas de spécialisation (tout le monde fait tout)

### Exemple Concret

Imaginons vendre un casque audio à $60 :
- Le client est VIP → devrait avoir -10%
- Concurrent le vend $55 → devrait baisser le prix
- Marge minimale 10% → ne peut pas aller sous $50

**Question** : Quel prix final ?

Un système monolithique mélange tout dans une formule complexe. Un système multi-agents **décompose** le problème.

## L'Approche Multi-Agents

### Principe : Divide & Conquer

Chaque agent a **une seule responsabilité** :

```
Agent 1 (Orchestrator) 
  ↓
  ├─→ Agent 2 (Customer Growth) : "Ce client vaut-il qu'on le garde ?"
  │   └─→ Output : shouldRetain=true, discount=10%
  │
  └─→ Agent 3 (Competitiveness) : "Quel prix pour gagner ?"
      └─→ Output : targetPrice=$55, position=2/5
  
Agent 1 synthétise : Prix final = $49.50 (VIP -10% + compétitif)
```

### Avantages

✅ **Séparation des préoccupations** : Chaque agent a un objectif clair  
✅ **Explicabilité** : On sait QUI a décidé QUOI  
✅ **Modularité** : Remplacer Agent 2 sans toucher Agent 3  
✅ **Spécialisation** : Chaque agent est expert de son domaine  
✅ **Testabilité** : Tester chaque agent indépendamment  

## Architecture 3-Agents en Détail

### Agent 1 : Vendor Orchestrator

**Rôle** : Chef d'orchestre, décision finale

**Inputs** :
- Produit (ID, prix de base)
- Client (email, historique)
- Context (time, stock, etc.)

**Logic** :
1. Consulter Agent 2 → recommandation fidélisation
2. Consulter Agent 3 → recommandation compétitivité
3. Synthétiser les deux
4. Appliquer garde-fous business

**Output** :
```go
type VendorDecision struct {
    FinalPrice    int64
    Discount      int64
    Margin        int
    Strategy      string
    Reasoning     []string
}
```

### Agent 2 : Customer Growth

**Rôle** : Maximiser customer lifetime value

**Inputs** :
- Customer ID
- Purchase history
- Engagement metrics

**Logic** :
- Tiers VIP (gold/silver/bronze)
- Should retain? (risque churn)
- Recommended discount

**Output** :
```go
type CustomerDecision struct {
    ShouldRetain bool
    Tier         string
    Discount     int  // percentage
    Reasoning    string
}
```

**Exemple** :
- Client gold, 5 achats ce mois → ShouldRetain=true, Discount=10%
- Client bronze, 0 achat → ShouldRetain=false, Discount=0%

### Agent 3 : Competitiveness

**Rôle** : Gagner face aux concurrents

**Inputs** :
- Product
- Our current price
- Competitor prices (via Shopping Graph)

**Logic** : Wraps 4 sub-agents :
1. **Price Intelligence** : Récupérer prix concurrents, calculer rank
2. **Market Analysis** : Tendance (up/down), opportunité
3. **Strategy Selection** : Choisir stratégie (premium/balanced/aggressive)
4. **Margin Validation** : Vérifier rentabilité

**Output** :
```go
type CompetitiveDecision struct {
    IsCompetitive  bool
    Position       string  // "1/5"
    TargetPrice    int64
    Strategy       string
    Confidence     int
}
```

## Pourquoi 3 Agents et Pas Plus ?

### On aurait pu avoir 2 agents
```
Agent 1 : Pricing
Agent 2 : Validation
```

**Problème** : Agent 1 mélange toujours fidélisation + compétition → pas assez granulaire.

### On aurait pu avoir 10 agents
```
Agent 1 : VIP Tier
Agent 2 : Discount Calculator
Agent 3 : Competitor Monitor
Agent 4 : Market Trends
Agent 5 : Margin Validator
...
```

**Problème** : Trop complexe à orchestrer, latence élevée, overhead.

### 3 = Sweet Spot

- **Agent 1** : Orchestration (responsabilité unique : décision finale)
- **Agent 2** : Dimension CLIENT (responsabilité unique : fidélisation)
- **Agent 3** : Dimension MARCHÉ (responsabilité unique : compétitivité)

Chaque dimension est indépendante et peut évoluer séparément.

## Le Pattern Wrapper : Agent 3 et ses 4 Sub-Agents

Agent 3 pourrait être monolithique, mais on le **décompose** en 4 étapes :

```
Agent 3 (Competitiveness)
  ├─→ 3.1 Price Intelligence   (QUOI : récupérer data)
  ├─→ 3.2 Market Analysis       (CONTEXTE : interpréter)
  ├─→ 3.3 Strategy Selection    (DÉCISION : choisir)
  └─→ 3.4 Margin Validation     (GARDE-FOU : valider)
```

Chaque sub-agent peut être **testé** et **modifié** indépendamment.

**Exemple** : Remplacer 3.1 pour utiliser une autre API de prix sans toucher 3.2/3.3/3.4.

## Comparaison avec d'Autres Patterns

### vs Rule-Based System

**Rule-Based** :
```python
if customer.tier == "gold":
    discount = 10
if competitor_price < our_price:
    our_price = competitor_price * 0.98
```

**Problème** : Ordre des règles ? Conflits ? Difficile à maintenir.

**Multi-Agent** : Chaque agent applique ses règles, orchestrator résout les conflits.

### vs Machine Learning

**ML** :
```python
price = model.predict([customer_features, market_features])
```

**Problème** : Black box, pas explicable, data dependency.

**Multi-Agent** : Logique explicite + peut intégrer ML dans un agent spécifique.

### vs Microservices

**Microservices** : Services indépendants (Customer Service, Pricing Service, Inventory Service).

**Multi-Agent** : Agents **collaboratifs** avec orchestration. Similaire mais focus sur décision vs. data.

## Pattern de Communication

### Séquentiel (Actuel)
```
Agent 1 → Agent 2 → attend résultat → Agent 3 → attend résultat → synthèse
```

**Avantage** : Simple, déterministe  
**Inconvénient** : Latence = somme des latences

### Parallèle (Possible)
```
Agent 1 → Agent 2 ↘
              → synthèse
Agent 1 → Agent 3 ↗
```

**Avantage** : Latence = max(Agent2, Agent3)  
**Inconvénient** : Agent 3 ne peut pas voir décision Agent 2

### Actuel = Séquentiel car Agent 1 passe résultat Agent 2 à Agent 3

```go
customerDecision := agent2.Analyze(req)
competitiveDecision := agent3.Analyze(req, customerDecision)  // ← dépendance
finalDecision := agent1.Synthesize(customerDecision, competitiveDecision)
```

## Behavior Emergent : La Guerre des Prix

**Comportement observé** : Les agents deviennent ultra-agressifs.

**Pourquoi ?**
1. Agent 3 voit concurrent à $55 → recommande $54
2. Concurrent voit notre $54 → recommande $53
3. Boucle jusqu'à marge négative

**Ce n'est pas un bug, c'est un comportement émergent** de l'interaction multi-agents.

**Solutions** :
- Marge minimale **stricte** (rejeter si < 10%)
- Cooldown (ne pas réagir instantanément)
- Mode "defensive" (protéger marge au lieu de gagner)

## Quand NE PAS Utiliser Multi-Agents

❌ **Pricing simple** : Prix fixe ou discount pourcentage → overkill  
❌ **Latence critique** : < 10ms requis → trop lent  
❌ **Pas de spécialisation** : Toutes les décisions similaires → un agent suffit  
❌ **Déterministe strict** : Besoin exactement le même prix à chaque fois  

## Quand Utiliser Multi-Agents

✅ **Décisions complexes** : Multiple dimensions (client, marché, inventaire)  
✅ **Explicabilité requise** : Besoin de tracer qui a décidé quoi  
✅ **Évolution fréquente** : Règles métier changent souvent  
✅ **Spécialisation** : Différentes expertises (data science, business, ops)  
✅ **Comportements adaptatifs** : Réponse au marché en temps réel  

## Conclusion

Le système multi-agents n'est **pas** une solution universelle, mais pour le pricing e-commerce compétitif :

- La **décomposition** en 3 agents (Vendor, Customer, Competitiveness) mappe bien au domaine métier
- L'**explicabilité** permet de comprendre et débugger les décisions
- La **modularité** permet d'évoluer chaque dimension indépendamment
- Les **comportements émergents** (guerre des prix) révèlent la complexité du domaine

Le trade-off principal : **complexité d'architecture** vs. **flexibilité métier**.

Pour UCP Merchant Test, ce trade-off est positif car le projet est une **plateforme d'expérimentation** où la flexibilité prime sur la simplicité.

## Lectures Complémentaires

- [ADR-0001 : Architecture Multi-Agents](../decisions/0001-architecture-multi-agents-pour-prix-competitif.md)
- [ADR-0004 : Architecture 3-Agents Orchestrée](../decisions/0004-architecture-3-agents-orchestree.md)
- [Reference : Agent Architecture](../reference/agent-architecture.md)
- [Explanation : Competitive Trade-offs](competitive-tradeoffs.md)
