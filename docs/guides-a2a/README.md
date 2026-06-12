# 📚 Guide Agents A2A - Version Simplifiée

**Architecture Microservices** pour démos rapides et POCs

**Pour les non-techniques** : Utilisez vos agents intelligents sans coder !

---

## 🎯 Par où commencer ?

### Jamais utilisé les agents ? → Tutorial

**[🚀 Mon premier agent](tutorial/premier-lancement.md)** (2 minutes)

Lancez votre premier agent intelligent et posez-lui une question. Aucune compétence technique requise.

---

## 📖 Comment faire... ? → How-to

Des recettes pratiques pour des tâches spécifiques :

| Guide | Temps | Niveau |
|-------|-------|--------|
| [Tester différents clients](how-to/tester-clients.md) | 5 min | ⭐ Facile |
| [Tester les prix des produits](how-to/tester-prix.md) | 5 min | ⭐ Facile |
| [Arrêter les agents](how-to/arreter-agents.md) | 1 min | ⭐ Facile |

---

## 🔍 Besoin d'une info précise ? → Reference

Tables de référence pour retrouver rapidement une information :

| Référence | Contenu |
|-----------|---------|
| [Clients disponibles](reference/clients-disponibles.md) | Liste des 4 clients de test (elsi, olwu, lja, manu) |
| [Produits disponibles](reference/produits-disponibles.md) | Liste des 4 produits (laptop, mouse, keyboard, monitor) |

---

## 💡 Pourquoi ça existe ? → Explanation

Comprendre les concepts derrière les agents :

| Guide | Objectif |
|-------|----------|
| [Pourquoi des agents indépendants ?](explanation/pourquoi-agents-independants.md) | Comprendre l'intérêt business et technique |

---

## 🗺️ Parcours recommandés

### Je veux juste tester (5 minutes)

1. [Mon premier agent](tutorial/premier-lancement.md) ← **Commencez ici**
2. [Tester différents clients](how-to/tester-clients.md)
3. [Arrêter les agents](how-to/arreter-agents.md)

### Je veux comprendre le concept (10 minutes)

1. [Pourquoi des agents indépendants ?](explanation/pourquoi-agents-independants.md)
2. [Mon premier agent](tutorial/premier-lancement.md)
3. [Tester les prix](how-to/tester-prix.md)

### Je prépare une démo client (15 minutes)

1. [Pourquoi des agents indépendants ?](explanation/pourquoi-agents-independants.md) ← Arguments business
2. [Mon premier agent](tutorial/premier-lancement.md) ← Mise en pratique
3. [Clients disponibles](reference/clients-disponibles.md) ← Données pour démo
4. [Produits disponibles](reference/produits-disponibles.md) ← Scénarios de prix

---

## ❓ Questions fréquentes

**Q : Je ne suis pas technique, est-ce pour moi ?**  
R : Oui ! Ce guide est fait pour vous. Tout se fait par copier/coller.

**Q : Combien de temps pour démarrer ?**  
R : 2 minutes chrono avec le [tutorial](tutorial/premier-lancement.md).

**Q : Ça fonctionne sur Mac et Windows ?**  
R : Oui, les deux sont supportés.

**Q : J'ai une erreur, que faire ?**  
R : Relancez `./scripts/stop-agents.sh` puis `./scripts/start-agents.sh`

**Q : Les données sont-elles réelles ?**  
R : Non, ce sont des données de test. Parfait pour des démos sans risque.

---

## 🆘 Besoin d'aide ?

Si vous êtes bloqué :

1. **Relancez tout** :
   ```bash
   ./scripts/stop-agents.sh
   ./scripts/start-agents.sh
   ```

2. **Fermez le terminal** et recommencez

3. **Contactez** l'équipe technique

---

## 📂 Structure de ce guide

```
guides/
├── README.md                              ← Vous êtes ici
├── tutorial/
│   └── premier-lancement.md               ← Démarrer en 2 min
├── how-to/
│   ├── tester-clients.md                  ← Recettes pratiques
│   ├── tester-prix.md
│   └── arreter-agents.md
├── reference/
│   ├── clients-disponibles.md             ← Tables de données
│   └── produits-disponibles.md
└── explanation/
    └── pourquoi-agents-independants.md    ← Concepts business
```

---

**Prêt à démarrer ?** → [Mon premier agent](tutorial/premier-lancement.md) 🚀
