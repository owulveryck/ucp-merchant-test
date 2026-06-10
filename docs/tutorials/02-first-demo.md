# Tutorial : Votre Première Démo Shopping Multi-Agents

**Durée estimée** : 20 minutes  
**Prérequis** : Tutorial 1 terminé, Vertex AI configuré (optionnel pour observer)

## Objectif

À la fin de ce tutorial, vous aurez :
- ✅ Lancé l'écosystème complet (3 merchants + Shopping Graph + Obs Hub)
- ✅ Observé un agent Gemini faire du shopping intelligent
- ✅ Compris comment fonctionne la recherche cross-merchant
- ✅ Vu un agent comparer les prix et choisir la meilleure offre

## Architecture de la démo

```
Client Agent (Gemini)
    ↓
Shopping Graph (:9000) ← recherche cross-merchant
    ↓
    ├─→ SuperShop  (:8182)
    ├─→ MegaMart   (:8183)
    └─→ BudgetBuy  (:8184)
         ↓
    Obs Hub (:9002) ← monitoring temps réel
```

## Étape 1 : Compiler les binaires de la démo

```bash
cd ~/stageocto/ucp-merchant-test

# Compiler tous les binaires
go build -o demo/bin/merchant ./sample_implementation
go build -o demo/bin/shopping-graph ./demo/cmd/shopping-graph
go build -o demo/bin/obs-hub ./demo/cmd/obs-hub
go build -o demo/bin/client ./demo/cmd/client

# Vérifier
ls -lh demo/bin/
```

**Résultat attendu** : 4 binaires compilés (merchant, shopping-graph, obs-hub, client).

## Étape 2 : Lancer les services (4 terminaux)

**Terminal 1 - Observability Hub** :
```bash
cd ~/stageocto/ucp-merchant-test
demo/bin/obs-hub --port 9002
```

**Terminal 2 - Shopping Graph** :
```bash
cd ~/stageocto/ucp-merchant-test
demo/bin/shopping-graph \
  --port 9000 \
  --config demo/config/shopping_graph.yaml \
  --obs-url http://localhost:9002
```

**Terminal 3 - Les 3 marchands** (script automatique) :
```bash
cd ~/stageocto/ucp-merchant-test
demo/scripts/start_merchants.sh
```

Ou manuellement dans 3 terminaux séparés :
```bash
# SuperShop
demo/bin/merchant --port 8182 --data-dir demo/data/merchant_a --data-format json --merchant-name SuperShop

# MegaMart
demo/bin/merchant --port 8183 --data-dir demo/data/merchant_b --data-format json --merchant-name MegaMart

# BudgetBuy
demo/bin/merchant --port 8184 --data-dir demo/data/merchant_c --data-format json --merchant-name BudgetBuy
```

## Étape 3 : Vérifier que tout tourne

```bash
# Shopping Graph
curl http://localhost:9000/health

# Merchants
curl http://localhost:8182/.well-known/ucp | jq .version
curl http://localhost:8183/.well-known/ucp | jq .version
curl http://localhost:8184/.well-known/ucp | jq .version

# Obs Hub
curl http://localhost:9002/health
```

**Résultat attendu** : Tous répondent 200 OK.

## Étape 4 : Tester la recherche cross-merchant

```bash
curl "http://localhost:9000/search?q=headphones" | jq
```

**Résultat attendu** : Liste de produits provenant des 3 marchands avec leurs prix.

## Étape 5 : Lancer l'agent client (optionnel - requiert Vertex AI)

**Configuration Vertex AI** :
```bash
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=us-central1
```

**Lancer le client** :
```bash
demo/bin/client \
  --graph-url http://localhost:9000 \
  --obs-url http://localhost:9002
```

**Exemple d'interaction** :
```
> Find me wireless headphones at the best price

Searching for "wireless headphones" across all merchants...

Found 3 results:
1. BudgetBuy - Wireless Audio Headphones - $79.99 (hints: BUDGET20)
2. MegaMart - Bluetooth Headset Pro - $84.99 (hints: MEGA10)
3. SuperShop - Wireless Headphones - $89.99 (hints: WELCOME15)

Creating checkouts...
Applying discounts...

Best price: BudgetBuy at $63.99 (after BUDGET20 discount)
Order placed! Order ID: ord_xxx
```

## Étape 6 : Observer dans le dashboard

Ouvrir http://localhost:9002 pour voir :
- Events en temps réel
- Décisions des agents
- Flux de recherche et checkout

## Ce que vous avez appris

✅ Architecture microservices (Shopping Graph + Merchants + Obs Hub)  
✅ Recherche cross-merchant avec Jaccard similarity  
✅ Agent autonome qui compare et optimise  
✅ Monitoring temps réel via SSE  

## Nettoyage

```bash
# Stopper tous les processus
pkill -f "demo/bin/"

# Ou individuellement
# Ctrl+C dans chaque terminal
```

## Prochaines étapes

- **Tutorial 3** : [Comprendre le pricing multi-agents](03-multi-agent-pricing.md)
- **How-to** : [Configurer un nouveau marchand](../how-to/configure-merchant.md)

## En cas de problème

**Les merchants ne se registrent pas dans Shopping Graph** :
- Vérifiez les URLs dans `demo/config/shopping_graph.yaml`
- Vérifiez les logs du Shopping Graph

**L'agent client ne répond pas** :
- Vérifiez les credentials Vertex AI : `gcloud auth application-default print-access-token`
- Vérifiez la variable `GOOGLE_CLOUD_PROJECT`

**Ports déjà utilisés** :
- Changez les ports avec les flags `--port`
- Libérez les ports : `lsof -ti:9000 | xargs kill`
