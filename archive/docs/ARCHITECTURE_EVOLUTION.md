# 📊 ARCHITECTURE & ÉVOLUTION DU PROJET

## 🎯 FICHIERS ACTIFS (Juin 4 - Version finale)

### Scripts principaux
- **arena_challenge.sh** ⭐ (Jun 4) - Démo challenge avec 4 concurrents
- **demo.sh** ⭐ (Jun 4) - Démo normale système 3-agents
- **acheter.sh** ⭐ (Jun 4) - Agent acheteur CLI

### Documentation actuelle
- **QUICK_START.md** (Jun 4) - Guide ultra-rapide
- **DEMO_MAITRE_STAGE.md** (Jun 4) - Script démo pour maître de stage
- **START_HERE.md** (Jun 4) - Point d'entrée
- **GUIDE_FINAL.md** (Jun 4) - Guide détaillé
- **CHALLENGE.md** (Jun 4) - Guide mode challenge

---

## 📚 HISTORIQUE PAR GÉNÉRATION

### GEN 1 - Mai 11-18 : Système 4-agents basique
**Objectif** : Pricing compétitif avec 4 agents autonomes

#### Scripts
- `demo_simple.sh` (Mai 18) - Premier test simple
- `run_competitive_demo.sh` (Mai 18) - Lancement basique
- `send_purchase_command.sh` (Mai 18) - Envoi commande CLI

#### Docs
- `COMPETITIVE_PRICING.md` (Mai 18) - Explication système 4-agents
- `DEMO_WEB.md` (Mai 12) - Première doc interface web
- `README.md` (Mai 11) - README original
- `CLAUDE.md` (Mai 11) - Architecture projet

**Garder ?** 
- ✅ `COMPETITIVE_PRICING.md` - Explique le système 4-agents sous-jacent
- ✅ `README.md` + `CLAUDE.md` - Doc projet
- ❌ Scripts - remplacés par versions plus récentes

---

### GEN 2 - Mai 28-29 : Agent acheteur + Observabilité
**Objectif** : Ajouter agent Gemini acheteur et dashboard d'observabilité

#### Scripts
- `run_multiagent_demo.sh` (Mai 28) - Avec agent Gemini
- `run_full_demo.sh` (Mai 28) - Démo complète avec obs
- `run_arena_demo.sh` (Mai 28) - Mode arène
- `stop_demo.sh` (Mai 28) - Arrêt services
- `test_auto_compete.sh` (Mai 28) - Test AUTO_COMPETE
- `test_multi_agent.sh` (Mai 28) - Test multi-agents

#### Docs
- `LAUNCH_MULTI_AGENT.md` (Mai 28) - Guide lancement
- `QUICKSTART_MULTIAGENT.md` (Mai 28) - Quick start multi-agents
- `README_DEMO.md` (Mai 28) - Guide démo
- `DEMO_SCENARIOS.md` (Mai 29) - Différents scénarios
- `DEMO.md` (Mai 28) - Doc démo générale
- `CHANGELOG.md` (Mai 28) - Historique changements

**Garder ?**
- ✅ `LAUNCH_MULTI_AGENT.md` - Explique l'intégration agent Gemini (contexte historique)
- ✅ `CHANGELOG.md` - Historique des changements (utile pour comprendre évolution)
- ⚠️ Scripts - fonctionnent mais nécessitent Gemini (compliqué à setup)
- ❌ Autres docs - remplacés par versions plus récentes

---

### GEN 3 - Juin 2-3 : Système 3-agents unifié
**Objectif** : Architecture 3-agents qui enveloppe le système 4-agents

#### Scripts
- `run_3agent_demo.sh` (Juin 2) - Premier test 3-agents
- `run_full_3agent_demo.sh` (Juin 2) - Démo complète 3-agents
- `run_unified_demo.sh` (Juin 3) - Version unifiée
- `demo_agents_simple.sh` (Juin 3) - Test simple
- `demo_complete.sh` (Juin 3) - Démo complète
- `demo_final.sh` (Juin 3) - Version finale (tentative)
- `test_auto_compete_reel.sh` (Juin 3) - Test réel AUTO_COMPETE
- `test_avec_sans_agents.sh` (Juin 3) - Comparaison avec/sans
- `test_demo_live.sh` (Juin 3) - Test live
- `test_quick.sh` (Juin 3) - Test rapide

#### Docs
- `README_MULTIAGENT.md` (Juin 3) - Guide multi-agents
- `README_SYSTEME_AGENTS.md` (Juin 3) - Système agents
- `COMMENT_TESTER.md` (Juin 3) - Guide test
- `DEMARRAGE.md` (Juin 3) - Guide démarrage
- `START_ICI.md` (Juin 3) - Point d'entrée
- `TEST_VALUES.md` (Juin 3) - Valeurs de test

**Garder ?**
- ✅ `run_unified_demo.sh` - Version intermédiaire fonctionnelle (référence évolution)
- ✅ `demo_complete.sh` - Démo complète avant version finale (backup)
- ❌ Autres scripts - remplacés par `arena_challenge.sh` / `demo.sh`
- ❌ Docs - remplacés par versions Juin 4

---

### GEN 4 - Juin 4 : Interface arène + Notifications
**Objectif** : Input arène, messages détaillés, toast notifications

#### Scripts actuels
- `arena_challenge.sh` ⭐ - Scénario challenge 4 concurrents
- `demo.sh` ⭐ - Mode flexible
- `acheter.sh` ⭐ - Agent CLI
- `init_marchand.sh` - Création marchands (obsolète, OAuth requis)
- `test_input_arena.sh` - Test input arène
- `test_demo_complete.sh` - Test complet

#### Docs actuelles
- `QUICK_START.md` ⭐
- `DEMO_MAITRE_STAGE.md` ⭐
- `START_HERE.md` ⭐
- `GUIDE_FINAL.md` ⭐
- `CHALLENGE.md` ⭐
- `ARENA_INPUT.md` - Doc input arène
- `FINAL_CHECK.md` - Checklist finale
- `TEST_NOW.md` - Guide test

**Garder ?**
- ✅ Tous les ⭐ (fichiers actifs)
- ❌ Scripts de test - développement terminé
- ❌ Docs intermédiaires - info dans guides principaux

---

## 🎯 RECOMMANDATION : STRUCTURE PROPRE

### À GARDER ABSOLUMENT (14 fichiers)

#### Scripts (3)
```
arena_challenge.sh          # Démo principale
demo.sh                     # Mode flexible
acheter.sh                  # Backup CLI
```

#### Docs actuelles (5)
```
QUICK_START.md              # Guide rapide
DEMO_MAITRE_STAGE.md        # Script démo
START_HERE.md               # Point d'entrée
GUIDE_FINAL.md              # Guide détaillé
CHALLENGE.md                # Guide challenge
```

#### Docs système (3)
```
README.md                   # README projet
CLAUDE.md                   # Architecture
COMPETITIVE_PRICING.md      # Système 4-agents sous-jacent
```

#### Historique (3)
```
CHANGELOG.md                # Évolution du projet
LAUNCH_MULTI_AGENT.md       # Contexte agent Gemini
DEMO_SCENARIOS.md           # Scénarios explorés
```

---

### À ARCHIVER (pour historique si besoin)

Créer un dossier `archive/` pour garder les versions intermédiaires :

```bash
mkdir -p archive/gen2-gemini archive/gen3-unified

# GEN 2 (Mai 28)
mv run_multiagent_demo.sh run_full_demo.sh run_arena_demo.sh archive/gen2-gemini/
mv test_multi_agent.sh test_auto_compete.sh archive/gen2-gemini/
mv QUICKSTART_MULTIAGENT.md README_DEMO.md DEMO.md archive/gen2-gemini/

# GEN 3 (Juin 2-3)
mv run_3agent_demo.sh run_full_3agent_demo.sh run_unified_demo.sh archive/gen3-unified/
mv demo_agents_simple.sh demo_complete.sh demo_final.sh archive/gen3-unified/
mv test_auto_compete_reel.sh test_avec_sans_agents.sh test_demo_live.sh archive/gen3-unified/
mv README_MULTIAGENT.md README_SYSTEME_AGENTS.md COMMENT_TESTER.md archive/gen3-unified/
mv DEMARRAGE.md START_ICI.md TEST_VALUES.md archive/gen3-unified/
```

---

### À SUPPRIMER (vraiment obsolètes)

#### Scripts cassés ou inutiles
```bash
rm demo_simple.sh              # Trop basique, remplacé
rm send_purchase_command.sh    # Une ligne, inutile
rm stop_demo.sh                # killall suffit
rm run_competitive_demo.sh     # Première version, obsolète
rm init_marchand.sh            # OAuth requis, ne marche pas
rm test_3_agents.sh            # Script minimal de test
rm test_quick.sh               # Remplacé par test_demo_complete.sh
rm test_input_arena.sh         # Test dev, plus nécessaire
rm test_demo_complete.sh       # Dev terminé
```

#### Docs redondantes
```bash
rm DEMO_WEB.md                 # Très ancienne, obsolète
rm ARENA_INPUT.md              # Info dans GUIDE_FINAL.md
rm FINAL_CHECK.md              # Info dans DEMO_MAITRE_STAGE.md
rm TEST_NOW.md                 # Dev terminé
rm CLEANUP_GUIDE.md            # Ce fichier remplace
```

---

## 📊 RÉSUMÉ

**Actuellement** : ~50 fichiers  
**Après nettoyage** : 14 fichiers + 20 archivés + suppression de 16

```
14 fichiers actifs (usage quotidien)
20 fichiers archivés (référence historique)
16 fichiers supprimés (vraiment obsolètes)
```

---

## 🚀 COMMANDE DE NETTOYAGE SÉCURISÉE

```bash
# 1. Créer archives
mkdir -p archive/gen2-gemini archive/gen3-unified

# 2. Archiver GEN 2
mv run_multiagent_demo.sh run_full_demo.sh run_arena_demo.sh \
   test_multi_agent.sh test_auto_compete.sh \
   QUICKSTART_MULTIAGENT.md README_DEMO.md DEMO.md \
   archive/gen2-gemini/

# 3. Archiver GEN 3
mv run_3agent_demo.sh run_full_3agent_demo.sh run_unified_demo.sh \
   demo_agents_simple.sh demo_complete.sh demo_final.sh \
   test_auto_compete_reel.sh test_avec_sans_agents.sh test_demo_live.sh \
   README_MULTIAGENT.md README_SYSTEME_AGENTS.md COMMENT_TESTER.md \
   DEMARRAGE.md START_ICI.md TEST_VALUES.md \
   archive/gen3-unified/

# 4. Supprimer obsolètes
rm -f demo_simple.sh send_purchase_command.sh stop_demo.sh \
      run_competitive_demo.sh init_marchand.sh test_3_agents.sh \
      test_quick.sh test_input_arena.sh test_demo_complete.sh \
      DEMO_WEB.md ARENA_INPUT.md FINAL_CHECK.md TEST_NOW.md \
      CLEANUP_GUIDE.md

echo "✅ Nettoyage terminé !"
echo "📁 $(ls -1 *.sh *.md 2>/dev/null | wc -l) fichiers actifs"
echo "📦 $(find archive -type f | wc -l) fichiers archivés"
```

---

**Avantages de cette approche :**
- ✅ Garde l'historique complet (dans archive/)
- ✅ Workspace propre (14 fichiers actifs)
- ✅ Évolution documentée (GEN 1 → 4)
- ✅ Facile de revenir en arrière si besoin
