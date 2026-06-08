# Tutorial : Démarrage Rapide

**Durée estimée** : 15 minutes  
**Prérequis** : Go 1.24+, Git

## Objectif

À la fin de ce tutorial, vous aurez :
- ✅ Compilé le serveur UCP merchant
- ✅ Lancé votre premier serveur avec le dataset flower shop
- ✅ Créé votre première checkout session
- ✅ Compris les bases de l'API UCP

## Étape 1 : Cloner et compiler

```bash
# Cloner le repository
cd ~/stageocto
git clone <repo-url> ucp-merchant-test
cd ucp-merchant-test

# Compiler le serveur
go build ./sample_implementation

# Vérifier la compilation
./sample_implementation --help
```

**Résultat attendu** : Vous devez voir l'aide avec les flags disponibles.

## Étape 2 : Lancer le serveur

```bash
# Lancer avec le dataset flower shop
./sample_implementation \
  --port 8182 \
  --data-dir sample_implementation/test_data/flower_shop \
  --simulation-secret super-secret-sim-key
```

**Résultat attendu** :
```
UCP Merchant Server starting on :8182
REST endpoint: http://localhost:8182/shopping-api
MCP endpoint: http://localhost:8182/mcp
Dashboard: http://localhost:8182/
```

## Étape 3 : Vérifier que ça marche

Ouvrir un nouveau terminal et tester la découverte UCP :

```bash
curl http://localhost:8182/.well-known/ucp | jq
```

**Résultat attendu** :
```json
{
  "version": "2026-01-11",
  "endpoints": {
    "shopping": "http://localhost:8182/shopping-api"
  }
}
```

## Étape 4 : Créer votre première checkout session

```bash
curl -X POST http://localhost:8182/shopping-api/checkout-sessions \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "item_id": "rose_bouquet",
        "quantity": 2
      }
    ]
  }' | jq
```

**Résultat attendu** : Une réponse JSON avec un `id` de session et le statut `incomplete`.

## Étape 5 : Ouvrir le dashboard

Ouvrez votre navigateur sur http://localhost:8182/

Vous devriez voir un dashboard avec :
- Liste des produits disponibles
- Inventaire en temps réel
- Events SSE (Server-Sent Events)

## Ce que vous avez appris

✅ Comment compiler et lancer le serveur UCP  
✅ Les 3 endpoints principaux (REST, MCP, Dashboard)  
✅ Comment créer une checkout session via l'API  
✅ Le concept de découverte UCP (`/.well-known/ucp`)

## Prochaines étapes

- **Tutorial 2** : [Lancer la démo shopping multi-agents](02-first-demo.md)
- **Tutorial 3** : [Comprendre le système de pricing intelligent](03-multi-agent-pricing.md)

## En cas de problème

**Le serveur ne démarre pas** :
- Vérifiez que le port 8182 est libre : `lsof -i :8182`
- Vérifiez Go version : `go version` (doit être >= 1.24)

**Les tests échouent** :
- Consultez [how-to/run-conformance-tests.md](../how-to/run-conformance-tests.md)

**Erreur sur le data-dir** :
- Vérifiez que le chemin existe et contient les CSV
- Le dataset flower shop est dans `sample_implementation/test_data/flower_shop/`
