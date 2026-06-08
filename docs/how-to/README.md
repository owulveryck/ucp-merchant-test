# How-to Guides - Accomplir des Tâches Spécifiques

Les how-to guides sont des **recettes** pour résoudre des problèmes concrets. Ils supposent que vous connaissez déjà les bases (voir [Tutorials](../tutorials/)).

## Guides Disponibles

### Configuration & Setup

#### [Configurer un Nouveau Marchand](configure-merchant.md)
**Problème résolu** : Je veux ajouter un 6ème marchand à l'Arena ou la démo shopping.

**Ce que vous allez faire** :
- Créer le répertoire de données
- Définir le catalog JSON
- Lancer le nouveau marchand
- L'intégrer au Shopping Graph

**Temps estimé** : 10 minutes

### Testing

#### [Exécuter les Tests de Conformance UCP](run-conformance-tests.md)
**Problème résolu** : Je veux vérifier que mon serveur passe tous les tests UCP.

**Ce que vous allez faire** :
- Cloner le repo conformance
- Lancer les 60 tests
- Débugger les échecs
- Intégrer en CI/CD

**Temps estimé** : 20 minutes

### Monitoring & Debugging

#### [Monitorer les Décisions Agents](monitor-agent-decisions.md)
**Problème résolu** : Je veux observer en temps réel ce que décident les agents.

**Ce que vous allez faire** :
- Utiliser le dashboard SSE
- Lire les logs détaillés
- Tracer une décision de pricing
- Identifier pourquoi un prix a été choisi

**Temps estimé** : 15 minutes

### Customization

#### [Configurer des Stratégies de Pricing Personnalisées](setup-pricing-strategies.md)
**Problème résolu** : Je veux modifier le comportement des agents (marges, stratégies, seuils).

**Ce que vous allez faire** :
- Modifier les seuils de marge
- Changer les stratégies par défaut
- Créer des règles par produit
- Activer le mode "defensive"

**Temps estimé** : 20 minutes

## Comment Utiliser ces Guides

1. **Identifiez votre problème** : Parcourez la liste ci-dessus
2. **Suivez les étapes** : Exécutez les commandes dans l'ordre
3. **Adaptez à votre cas** : Les guides montrent un exemple, ajustez les valeurs
4. **Vérifiez le résultat** : Chaque guide indique comment valider le succès

## Différence avec Tutorials

| Tutorial | How-to |
|----------|--------|
| Apprendre un concept | Résoudre un problème |
| Parcours complet | Tâche précise |
| Explications détaillées | Instructions concises |
| Pour débutants | Pour utilisateurs confirmés |

**Exemple** :
- **Tutorial** : "Comprendre le pricing multi-agents" → vous fait découvrir le système
- **How-to** : "Configurer une stratégie de pricing" → vous montre comment changer un paramètre

## Guides à Venir

- [ ] Déployer en Production avec Docker
- [ ] Configurer des Webhooks Custom
- [ ] Intégrer un Nouveau Transport (gRPC)
- [ ] Créer un Agent Custom
- [ ] Optimiser les Performances
- [ ] Setup Monitoring avec Prometheus
- [ ] Configurer Multi-Currency

## Besoin d'un Guide qui N'Existe Pas ?

Ouvrez une issue sur GitHub avec :
- Le problème que vous voulez résoudre
- Ce que vous avez déjà essayé
- Le résultat attendu

## Contribuer aux How-to Guides

Les how-to doivent être :
- **Orientés tâche** : Résoudre un problème précis
- **Concis** : Pas d'explications longues, juste les étapes
- **Testés** : Vérifiez que les commandes marchent
- **Complets** : De l'état initial au résultat final

Structure recommandée :
1. Contexte (1-2 phrases)
2. Prérequis
3. Étapes numérotées
4. Vérification du résultat
5. Troubleshooting commun

Pas d'explication du "pourquoi" → ça va dans [Explanation](../explanation/).
