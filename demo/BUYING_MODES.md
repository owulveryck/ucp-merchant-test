# Modes d'achat de l'Agent Acheteur

L'agent acheteur Gemini supporte **2 modes d'optimisation** détectés automatiquement selon votre instruction.

---

## 💰 Mode "Moins cher" (par défaut)

**Déclenché par** : instructions classiques sans mention de rapidité
- "Achète des fleurs"
- "Trouve le meilleur prix"
- "Je veux des roses pas chères"

**Comportement** :
1. Sélectionne l'**option de livraison la moins chère** chez chaque marchand
2. Compare les **totaux finaux** (prix + livraison - remises)
3. Achète chez le marchand avec le **prix total le plus bas**

**Exemple** :
```
Marchand A : $65 + $5 livraison standard = $70
Marchand B : $70 + $0 livraison gratuite = $70
Marchand C : $75 + $10 express = $85

→ Choix : A ou B ($70)
```

---

## ⚡ Mode "Plus rapide"

**Déclenché par** : mots-clés de rapidité
- "Achète des fleurs **rapidement**"
- "Livraison **express**"
- "J'en ai besoin **vite**"
- Français : rapide, vite, urgent
- Anglais : fast, quick, asap, express, urgent

**Comportement** :
1. Sélectionne l'**option de livraison la plus rapide** chez chaque marchand
   - Express (1-2 jours) > Standard (3-5 jours)
2. Compare les **délais de livraison**
3. Achète chez le marchand avec le **délai le plus court**

**Exemple** :
```
Marchand A : Standard (5 jours) - $70 total
Marchand B : Standard (3 jours) - $70 total
Marchand C : Express (1 jour) - $85 total

→ Choix : C (1 jour, même si +$15)
```

---

## Détection des délais

L'agent estime les délais depuis les titres d'options de shipping :

| Titre contient | Délai estimé |
|----------------|--------------|
| "Express", "Expedited" | 1-2 jours |
| "Standard", "Regular" | 3-5 jours |

---

## Algorithme de ranking (Shopping Graph)

Le Shopping Graph utilise l'algorithme **Arena** par défaut pour classer les résultats de recherche.

### Score de qualité (0-10 points)

```
Score = Prix (8 pts) + Stock (2 pts)
```

- **Prix** : inversement proportionnel (prix bas = +8 pts, prix élevé = moins)
- **Stock** : +2 pts si disponible, 0 si rupture

### Ce qui est EXCLU du score

❌ **Bid (enchères publicitaires)** : n'affecte que l'**ordre d'affichage** (résultats sponsorisés en premier), PAS le score qualité.

Le bid est pour la **visibilité**, pas pour la **valeur client**.

---

## Résumé

| Critère | Moins cher | Plus rapide |
|---------|------------|-------------|
| **Optimise** | Prix total | Délai de livraison |
| **Shipping** | Option la moins chère | Option la plus rapide |
| **Sélection finale** | Total le plus bas | Délai le plus court |
| **Accepte surcoût** | ❌ Non | ✅ Oui (pour gagner du temps) |

---

## Pour la démo

**Instructions à tester** :

```
💰 Moins cher :
- "Achète un bouquet de roses"
- "Trouve le meilleur prix pour des fleurs"

⚡ Plus rapide :
- "Achète un bouquet rapidement"
- "J'ai besoin de fleurs en express"
- "Livraison rapide SVP"
```

L'agent annoncera son mode au démarrage :
```
Instruction: Achète des fleurs rapidement (merchants: 3, mode: fastest)
```
