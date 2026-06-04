# 📋 ADRs À CRÉER

Basé sur les décisions architecturales prises en Juin 2026.

## 🆕 ADRs Nouvelles Décisions (Juin 2026)

### ADR-004 : Architecture 3-Agents Enveloppant le Système 4-Agents

**Décision** : Créer une architecture 3-agents qui **enveloppe** le système 4-agents existant au lieu de le remplacer

**Problème** :
- Système 4-agents fonctionne mais n'est pas orchestré
- Pas de couche pour la croissance client (fidélité, tiers)
- Pas de coordination entre analyse client et analyse compétitive

**Solution choisie** :
- **Agent 1 (Vendeur/Orchestrateur)** : Coordonne Agent 2 et 3, décision finale
- **Agent 2 (Customer Growth)** : Analyse profil client, calcule bonus fidélité (0-15%)
- **Agent 3 (Compétitivité)** : **Enveloppe les 4 agents existants**, interroge Shopping Graph

**Alternatives considérées** :
1. Remplacer le système 4-agents → Rejeté (perte de logique éprouvée)
2. Ajouter les 2 nouveaux agents en parallèle → Rejeté (pas d'orchestration)
3. Architecture hybride (enveloppe) → **Choisi**

**Conséquences** :
- ✅ Réutilise logique 4-agents éprouvée
- ✅ Ajoute couche client et orchestration
- ✅ Extensible (facile d'ajouter Agent 4, 5...)
- ❌ Complexité (7 agents au total)

**Fichiers** : `pkg/pricing-unified/`, `demo/cmd/arena/tenant_3agents.go`

---

### ADR-005 : Interface Arène avec Agent Acheteur Intégré

**Décision** : Intégrer l'agent acheteur directement dans l'interface web au lieu d'un agent externe Gemini

**Problème** :
- Agent Gemini externe nécessite GCP, API keys, setup complexe
- Utilisateur doit lancer un script séparé (`acheter.sh`)
- Pas de feedback visuel en temps réel
- Difficile à démontrer

**Solution choisie** :
- Input intégré dans dashboard arène
- Fonction `executeBuyingFlow()` côté serveur (Go)
- Messages SSE en temps réel
- Toast notifications + surbrillance marchand gagnant

**Alternatives considérées** :
1. Garder agent Gemini externe → Rejeté (setup complexe)
2. Agent côté client (JavaScript) → Rejeté (pas d'accès Shopping Graph)
3. Agent intégré serveur → **Choisi**

**Conséquences** :
- ✅ Zero setup (pas besoin GCP/Gemini)
- ✅ Feedback temps réel (SSE)
- ✅ UX améliorée (notifications toast)
- ✅ Facile à démontrer
- ❌ Moins "intelligent" qu'un vrai LLM
- ❌ Logique achat simple (pas de NLP)

**Fichiers** : `demo/internal/obs/handler.go` (executeBuyingFlow), `demo/internal/obs/dashboard_arena.go` (UI)

---

### ADR-006 : Messages Détaillés de Décision d'Achat

**Décision** : Afficher un raisonnement détaillé lors de la sélection du marchand gagnant

**Problème** :
- L'agent acheteur choisit un marchand mais on ne sait pas pourquoi
- Pas de transparence sur la comparaison des prix
- Difficile de valider que le système fonctionne correctement

**Solution choisie** :
- Message de **comparaison des prix** (tous les marchands avec écarts)
- Message de **décision** avec justification :
  - Prix le plus bas
  - Nombre de concurrents
  - Économie vs 2ème meilleur prix (%)
- Message de **confirmation d'achat**

**Exemple** :
```
📊 Comparaison des prix :
   • MonMagasin: $51.30 ← ✅ LE MOINS CHER
   • PrixCassés: $58.00 (+$6.70)
   • TopPrix: $59.00 (+$7.70)

🎯 DÉCISION : MonMagasin sélectionné !
Pourquoi ?
   • Prix le plus bas : $51.30
   • 4 concurrent(s) comparé(s)
   • Économie : 11.6% vs 2ème meilleur prix

✅ Achat confirmé ! Prix final: $51.30
```

**Alternatives considérées** :
1. Message simple "X gagne" → Rejeté (pas de transparence)
2. Seulement le prix final → Rejeté (pas de justification)
3. Comparaison + décision détaillée → **Choisi**

**Conséquences** :
- ✅ Transparence totale
- ✅ Validation facile du système
- ✅ Démo plus convaincante
- ❌ Plus de messages (peut être verbeux)

**Fichiers** : `demo/internal/obs/handler.go` (lignes 349-410)

---

### ADR-007 : Toast Notifications pour Actions Importantes

**Décision** : Afficher des notifications toast pour les événements critiques (marchand sélectionné)

**Problème** :
- Le panel d'activité peut être fermé
- Messages dans logs difficiles à repérer
- Pas de feedback visuel immédiat

**Solution choisie** :
- Toast notifications en haut à droite
- Auto-dismiss après 10s
- Animations slide-in/slide-out
- Couleurs selon type (success = vert)

**Alternatives considérées** :
1. Seulement logs dans panel → Rejeté (pas assez visible)
2. Modal bloquante → Rejeté (intrusive)
3. Toast non-intrusive → **Choisi**

**Conséquences** :
- ✅ Feedback visuel immédiat
- ✅ Non-intrusif (auto-dismiss)
- ✅ UX moderne
- ❌ Code CSS/JS supplémentaire

**Fichiers** : `demo/internal/obs/dashboard_arena.go` (CSS toast + JS)

---

### ADR-008 : Scénario Challenge avec Concurrents Pré-Créés

**Décision** : Créer un script `arena_challenge.sh` qui lance 4 concurrents automatiquement

**Problème** :
- Démo normale nécessite création manuelle de 3-4 marchands
- Temps de setup long (2-3 minutes)
- Pas de scénario "perdant → gagnant" clair

**Solution choisie** :
- Script qui crée automatiquement 4 concurrents avec prix fixes
- Utilisateur arrive en "outsider" perdant
- Active système 3-agents → devient gagnant
- Scénario narratif clair

**Alternatives considérées** :
1. Seulement `demo.sh` (création manuelle) → Rejeté (setup long)
2. Auto-création au démarrage → Rejeté (pas de contrôle)
3. Script challenge séparé → **Choisi**

**Conséquences** :
- ✅ Démo rapide (30 secondes vs 3 minutes)
- ✅ Scénario dramatique (underdog story)
- ✅ Reproductible
- ❌ Moins flexible que création manuelle

**Fichiers** : `scripts/arena_challenge.sh`

---

## 🔄 ADRs À TRADUIRE EN FRANÇAIS

Les ADRs existants sont en anglais, il faut créer versions françaises :

### À faire :
1. ✅ ADR-001 : Déjà traduit (`0001-architecture-multi-agents-pour-prix-competitif.md`)
2. ✅ ADR-002 : Déjà traduit (`0002-strategie-victoire-avant-marge-parfaite.md`)
3. ✅ ADR-003 : Déjà traduit (`0003-strategie-detection-codes-promo.md`)
4. ❌ `docs/adr/001-multi-agent-shopping-architecture.md` → À traduire
5. ❌ `docs/adr/002-multi-transport-architecture.md` → À traduire (déjà reformaté MADR)
6. ❌ `docs/adr/003-competitive-pricing-agent.md` → À traduire

---

## 📊 PRIORITÉS

### Haute Priorité (Décisions Juin 2026 - Pour démo)
1. **ADR-004** : Architecture 3-agents (décision principale)
2. **ADR-005** : Agent acheteur intégré (innovation UX)

### Moyenne Priorité (Améliorations UX)
3. **ADR-006** : Messages détaillés
4. **ADR-007** : Toast notifications
5. **ADR-008** : Scénario challenge

### Basse Priorité (Traductions)
6. Traduire ADR-001, 002, 003 des `docs/adr/` en français

---

## 🎯 RECOMMANDATION

**Créer d'abord ADR-004** (Architecture 3-agents) car c'est :
- La décision architecturale majeure
- Ce qui différencie ton travail de stage
- Ce que ton maître de stage veut voir

**Format** : Suivre template MADR comme ADR-002 reformaté
**Langue** : Français (pour cohérence avec ADRs `docs/decisions/`)
**Emplacement** : `docs/decisions/0004-architecture-3-agents-orchestree.md`

Veux-tu que je crée ADR-004 maintenant ? 📝
