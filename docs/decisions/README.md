# Architecture Decision Records (ADRs)

Ce répertoire contient les Architecture Decision Records (ADRs) pour le système UCP Merchant Test - Competitive Pricing Intelligence.

## Index des ADRs

### 🇫🇷 Décisions en Français (Principales)

| ADR | Titre | Statut | Date |
|-----|-------|--------|------|
| [0001](0001-architecture-multi-agents-pour-prix-competitif.md) | Architecture Multi-Agents pour Pricing Compétitif | Accepté | 2026-05-29 |
| [0002](0002-strategie-victoire-avant-marge-parfaite.md) | Stratégie Victoire Avant Marge Parfaite | Accepté | 2026-05-29 |
| [0003](0003-strategie-detection-codes-promo.md) | Stratégie de Détection des Codes Promo | Accepté | 2026-05-29 |
| [0004](0004-architecture-3-agents-orchestree.md) | Architecture 3-Agents Orchestrée (Juin 2026) | Accepté | 2026-06-04 |
| [0005](0005-agent-acheteur-integre.md) | Agent Acheteur Intégré dans Interface Web | Accepté | 2026-06-04 |
| [0006](0006-messages-detailles-decision-achat.md) | Messages Détaillés de Décision d'Achat | Accepté | 2026-06-04 |
| [0007](0007-scenario-challenge-concurrents.md) | Scénario Challenge avec Concurrents Pré-Créés | Accepté | 2026-06-04 |
| [0008](0008-multi-agent-shopping-architecture.md) | Multi-Agent Shopping Architecture | Accepté | 2026-05-29 |
| [0009](0009-multi-transport-architecture.md) | Architecture Multi-Transport (REST/MCP/A2A) | Accepté | 2026-01-11 |

### 🇬🇧 English Versions (Reference)

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0010](0010-multi-agent-architecture-for-competitive-pricing.md) | Multi-Agent Architecture for Competitive Pricing | Accepted | 2026-05-29 |
| [0011](0011-winning-strategy-over-perfect-margin.md) | Winning Strategy Over Perfect Margin | Accepted | 2026-05-29 |
| [0012](0012-competitive-pricing-agent.md) | Competitive Pricing Agent | Accepted | 2026-05-29 |
| [0013](0013-discount-code-detection-strategy.md) | Discount Code Detection Strategy | Accepted | 2026-05-29 |

## Organisation

**ADRs 0001-0003** : Système 4-agents initial (Mai 2026)
- Architecture multi-agents de base
- Stratégie de pricing compétitif
- Détection codes promo

**ADRs 0004-0007** : Évolution 3-agents + UX (Juin 2026)
- Architecture 3-agents enveloppant le système 4-agents
- Agent acheteur intégré avec feedback temps réel
- Messages détaillés pour transparence
- Scénario de démo challenge

**ADRs 0008-0009** : Infrastructures
- Architecture shopping multi-agents
- Support multi-transport (REST/MCP/A2A)

**ADRs 0010-0013** : Versions anglaises de référence

## Décisions Clés

### 🎯 ADR-0004 : Architecture 3-Agents (Juin 2026)

**Problème** : Système 4-agents fonctionne mais manque d'orchestration et d'analyse client

**Décision** : Architecture hybride où 3 nouveaux agents enveloppent le système 4-agents existant
- Agent 1 (Vendeur) : Orchestrateur
- Agent 2 (Customer Growth) : Analyse client, fidélité
- Agent 3 (Compétitivité) : Enveloppe les 4 agents existants

**Impact** : Passage de 0% à 100% de ventes (prix $70 → $51.30)

---

### 🛒 ADR-0005 : Agent Acheteur Intégré (Juin 2026)

**Problème** : Agent Gemini externe nécessite setup complexe (GCP, API keys)

**Décision** : Agent intégré côté serveur (Go) avec SSE temps réel

**Bénéfice** : Zero setup, feedback visuel instantané, facile à démontrer

---

### 📊 ADR-0006 : Messages Détaillés (Juin 2026)

**Problème** : Décision de l'agent acheteur opaque, difficile à valider

**Décision** : Afficher comparaison complète des prix + justification chiffrée

**Impact** : Transparence totale, validation immédiate, démo convaincante

---

### 🏆 ADR-0007 : Scénario Challenge (Juin 2026)

**Problème** : Démo manuelle nécessite 3-5 minutes de setup

**Décision** : Script `arena_challenge.sh` crée automatiquement 4 concurrents

**Impact** : Setup 30 secondes, scénario "perdant → gagnant" dramatique

---

## Statuts ADR

- **Accepté** : Décision approuvée et implémentée
- **Proposé** : En cours de révision
- **Déprécié** : Plus actuel mais conservé pour historique
- **Remplacé** : Supplanté par un ADR plus récent
- **Rejeté** : Considéré mais non retenu

## Format

Tous les ADRs suivent le template [MADR](https://adr.github.io/madr/) :
- Contexte et Problème
- Facteurs de Décision
- Options Considérées
- Décision avec Conséquences
- Avantages/Inconvénients des Options

## Documentation Associée

- [QUICK_START.md](../../QUICK_START.md) : Démarrage rapide
- [README_ADR.md](README_ADR.md) : Guide détaillé ADRs (legacy)
