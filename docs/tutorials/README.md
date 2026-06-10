# Tutorials - Apprendre en Pratiquant

Les tutorials sont conçus pour les **débutants** qui découvrent le projet. Chaque tutorial est un guide pas-à-pas qui vous emmène de zéro à un résultat concret.

## Parcours d'Apprentissage

### 1. [Getting Started](01-getting-started.md)
**Durée** : 15 minutes  
**Niveau** : Débutant  
**Vous allez apprendre** :
- Compiler et lancer le serveur UCP
- Créer votre première checkout session
- Comprendre les endpoints REST/MCP
- Utiliser le dashboard SSE

**Commencez ici** si c'est votre première fois sur le projet.

### 2. [Votre Première Démo Shopping](02-first-demo.md)
**Durée** : 20 minutes  
**Niveau** : Débutant  
**Prérequis** : Tutorial 1 terminé

**Vous allez apprendre** :
- Lancer l'écosystème complet (3 merchants + Shopping Graph + Obs Hub)
- Observer un agent Gemini faire du shopping intelligent
- Comprendre la recherche cross-merchant
- Voir un agent comparer les prix et optimiser

**Faites ce tutorial** pour comprendre l'architecture shopping multi-agents.

### 3. [Comprendre le Pricing Multi-Agents](03-multi-agent-pricing.md)
**Durée** : 30 minutes  
**Niveau** : Intermédiaire  
**Prérequis** : Tutorials 1 & 2 terminés

**Vous allez apprendre** :
- Lancer l'Arena avec 5 marchands en compétition
- Comprendre le système 3-agents (Vendor, Customer Growth, Competitiveness)
- Observer une guerre des prix en temps réel
- Analyser les décisions via logs et dashboard

**Faites ce tutorial** pour comprendre le cœur du système de pricing intelligent.

## Progression Recommandée

```
Tutorial 1 (Getting Started)
    ↓
Tutorial 2 (Shopping Demo)
    ↓
Tutorial 3 (Multi-Agent Pricing)
    ↓
[How-to Guides] pour des tâches spécifiques
```

## Après les Tutorials

Une fois les 3 tutorials terminés, vous devriez :
- ✅ Savoir compiler et lancer tous les binaires
- ✅ Comprendre l'architecture 3-couches (UCP Base / Shopping Demo / Arena)
- ✅ Avoir observé les agents en action
- ✅ Pouvoir lire les logs et comprendre les décisions

**Prochaines étapes** :
- Consultez [How-to Guides](../how-to/) pour accomplir des tâches spécifiques
- Consultez [Reference](../reference/) pour les détails techniques
- Consultez [Explanation](../explanation/) pour comprendre le "pourquoi"

## Besoin d'Aide ?

Si vous bloquez sur un tutorial :
1. Vérifiez la section "En cas de problème" à la fin du tutorial
2. Consultez [Reference: Error Codes](../reference/error-codes.md)
3. Regardez les logs : `tail -f logs/*.log`
4. Ouvrez une issue sur GitHub

## Contribuer aux Tutorials

Les tutorials doivent être :
- **Orientés apprentissage** : Pour découvrir, pas pour résoudre un problème
- **Pas-à-pas** : Chaque étape claire avec résultat attendu
- **Testés** : Vérifiez que les commandes fonctionnent réellement
- **Concrets** : Aboutir à un résultat tangible (serveur qui tourne, démo qui marche)

Ne mélangez pas tutorial et how-to. Un tutorial enseigne, un how-to résout.
