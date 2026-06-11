# 📚 Guide Arena - Version Simplifiée

**Architecture Monolithe** pour production haute performance

**Pour les non-techniques** : Comprenez comment 4 agents intelligents travaillent ensemble !

---

## 🎯 Par où commencer ?

### Jamais utilisé l'Arena ? → Tutorial

**[🚀 Mon premier lancement](tutorial/premier-lancement.md)** (5 minutes)

Lancez l'environnement Arena et voyez les 4 agents travailler ensemble. Aucune compétence technique requise.

---

## 📖 Comment faire... ? → How-to

Des recettes pratiques pour utiliser l'Arena :

| Guide | Temps | Niveau |
|-------|-------|--------|
| [Voir les logs de décision](how-to/voir-logs.md) | 2 min | ⭐ Facile |
| [Arrêter l'Arena](how-to/arreter-arena.md) | 1 min | ⭐ Facile |

---

## 🔍 Besoin d'une info précise ? → Reference

Tables de référence sur les agents :

| Référence | Contenu |
|-----------|---------|
| [Les 4 agents Arena](reference/agents-arena.md) | Détails techniques de chaque agent (fonctions, temps, règles) |

---

## 💡 Pourquoi ça existe ? → Explanation

Comprendre l'évolution de l'Arena :

| Guide | Objectif |
|-------|----------|
| [Avant/Après : L'évolution](explanation/avant-apres.md) | Comparaison Arena basique vs Arena intelligente |
| [Les 4 agents intelligents](explanation/les-4-agents.md) | Rôle et collaboration des agents |

---

## 🗺️ Parcours recommandés

### Je veux juste tester (10 minutes)

1. [Mon premier lancement](tutorial/premier-lancement.md) ← **Commencez ici**
2. [Voir les logs de décision](how-to/voir-logs.md)
3. [Arrêter l'Arena](how-to/arreter-arena.md)

### Je veux comprendre le concept (15 minutes)

1. [Avant/Après : L'évolution](explanation/avant-apres.md) ← Arguments business
2. [Les 4 agents intelligents](explanation/les-4-agents.md)
3. [Mon premier lancement](tutorial/premier-lancement.md)

### Je prépare une démo client (20 minutes)

1. [Avant/Après : L'évolution](explanation/avant-apres.md) ← Pourquoi 4 agents ?
2. [Les 4 agents intelligents](explanation/les-4-agents.md) ← Comment ils travaillent
3. [Mon premier lancement](tutorial/premier-lancement.md) ← Démonstration live
4. [Voir les logs de décision](how-to/voir-logs.md) ← Prouver l'intelligence

---

## ❓ Questions fréquentes

**Q : Quelle différence avec les Agents A2A ?**  
R : Arena = Production (4 agents monolithe), A2A = Démos rapides (agents séparés). [Voir comparaison](explanation/avant-apres.md)

**Q : Combien de temps pour démarrer ?**  
R : 5 minutes avec le [tutorial](tutorial/premier-lancement.md).

**Q : Les agents sont-ils vraiment intelligents ?**  
R : Oui ! Intelligence décisionnelle basée sur données réelles. [Voir détails](explanation/les-4-agents.md)

**Q : Quel est le taux de victoire ?**  
R : 78% avec 4 agents vs 45% avant. [Voir statistiques](explanation/avant-apres.md)

**Q : Les marges sont-elles garanties ?**  
R : Oui, le Contrôleur valide que chaque vente est rentable (≥10%). [Voir référence](reference/agents-arena.md)

---

## 📊 Résultats clés

**Performance** :
- 🏆 Taux de victoire : **78%** (vs 45% avant)
- 💰 Marges : **≥10%** garanties
- ⚡ Temps de décision : **140ms**
- 📈 Profit : **+151%** vs avant

**Les 4 agents** :
- 🕵️ **L'Espion** : Récupère les prix concurrents
- 📊 **L'Analyste** : Évalue la position marché
- 🎯 **Le Stratège** : Recommande la stratégie
- ✅ **Le Contrôleur** : Valide la rentabilité

---

## 🆘 Besoin d'aide ?

Si vous êtes bloqué :

1. **Relancez tout** :
   ```bash
   ./scripts/stop-demo.sh
   ./scripts/start-demo.sh
   ```

2. **Vérifiez les logs** :
   ```bash
   tail -f logs/arena.log
   ```

3. **Contactez** l'équipe technique

---

## 🔗 Liens utiles

- **[Agents A2A](../guides-a2a/README.md)** : Architecture microservices pour démos
- **[ADRs Arena](../decisions/arena/)** : Décisions architecture détaillées
- **[Navigation générale](../NAVIGATION.md)** : Guide complet du repository

---

## 📂 Structure de ce guide

```
guides-arena/
├── README.md                              ← Vous êtes ici
├── tutorial/
│   └── premier-lancement.md               ← Démarrer en 5 min
├── how-to/
│   ├── voir-logs.md                       ← Recettes pratiques
│   └── arreter-arena.md
├── reference/
│   └── agents-arena.md                    ← Spécifications techniques
└── explanation/
    ├── avant-apres.md                     ← Évolution Arena
    └── les-4-agents.md                    ← Comment ils travaillent
```

---

**Prêt à démarrer ?** → [Mon premier lancement](tutorial/premier-lancement.md) 🚀
