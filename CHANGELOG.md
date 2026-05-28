# Changelog

## [v1.0-multi-agent-discounts] - 2026-05-28

### ✨ Nouvelles fonctionnalités

- **Multi-agent architecture** : 4 agents spécialisés pour le pricing compétitif
  - Agent 1 (Espion) : Collecte les prix des concurrents
  - Agent 2 (Analyste) : Analyse la position sur le marché
  - Agent 3 (Stratège) : Recommande une stratégie de prix
  - Agent 4 (Contrôleur) : Valide les marges

- **Interface utilisateur simplifiée**
  - Un seul bouton "💡 Calculer le meilleur prix"
  - Section "Intelligence de Prix" unifiée
  - Raisonnement des agents affiché en français simple
  - Pas besoin de connaissances techniques

- **Prise en compte des codes promo concurrents**
  - Détecte les codes promo disponibles (WELCOME10, SAVE20, etc.)
  - Calcule le prix effectif après réduction
  - Recommande un prix qui bat le VRAI prix concurrent
  - Parsing intelligent : WELCOME10 → -10%, FIXED500 → -$5

### 🔧 Améliorations techniques

- Architecture simplifiée : réponse HTTP directe (pas de SSE)
- Tags JSON ajoutés sur SearchResult pour unmarshaling correct
- GetLastDecisions() pour récupérer les décisions des agents
- estimateEffectivePrice() pour calculer les prix après promo

### 🐛 Corrections de bugs

- **Fix**: Les agents ne s'affichaient pas (manquait flag --competitive-pricing)
- **Fix**: Prix concurrent incorrect (manquait tags JSON sur SearchResult)
- **Fix**: Prix recommandé ne s'affichait pas (complexité SSE remplacée par HTTP direct)
- **Fix**: MonMagasin perdait face aux codes promo (agents ignoraient les discounts)

### 📝 Commits principaux

- `c6fbae5` - FIX: Agent considers competitor discounts when calculating lowest price
- `99b6867` - ADD: Return agent reasoning in API response for dashboard display
- `34d5b1b` - SIMPLIFY: Replace SSE callback with direct HTTP response
- `69fab2b` - FIX: Add missing JSON tags to SearchResult struct

### 🚀 Utilisation

```bash
./run_full_demo.sh
```

Puis dans le dashboard :
1. Créez 2-3 marchands avec des prix différents
2. Ajoutez des codes promo sur les concurrents
3. Cliquez "💡 Calculer le meilleur prix"
4. Vérifiez le raisonnement des 4 agents
5. Cliquez "Appliquer ce prix"
6. Lancez un agent acheteur

MonMagasin devrait gagner contre les concurrents même avec leurs codes promo ! 🏆
