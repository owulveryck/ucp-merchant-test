# How-to : Exécuter les Tests de Conformance UCP

## Contexte

Le projet doit passer **60 tests UCP** répartis dans 13 fichiers. Ce guide explique comment les exécuter.

## Prérequis

- Serveur merchant compilé
- Python 3.8+ installé
- Repository UCP conformance cloné

## Cloner le repository de conformance

```bash
cd ~/projects
git clone https://github.com/universal-commerce-protocol/conformance.git
cd conformance
```

## Lancer le serveur merchant

**Terminal 1** :
```bash
cd ~/stageocto/ucp-merchant-test
go run ./sample_implementation \
  --port 8182 \
  --data-dir ~/projects/conformance/test_data/flower_shop \
  --simulation-secret super-secret-sim-key
```

## Exécuter tous les tests

**Terminal 2** :
```bash
cd ~/projects/conformance

# Méthode 1 : Script automatique
for test_file in *_test.py; do
  echo "Running $test_file..."
  python3 "$test_file" \
    --server_url=http://localhost:8182 \
    --simulation_secret=super-secret-sim-key \
    --conformance_input=test_data/flower_shop/conformance_input.json \
    --test_data_dir=test_data/flower_shop
done
```

## Exécuter un test spécifique

```bash
cd ~/projects/conformance

python3 checkout_lifecycle_test.py \
  --server_url=http://localhost:8182 \
  --simulation_secret=super-secret-sim-key \
  --conformance_input=test_data/flower_shop/conformance_input.json \
  --test_data_dir=test_data/flower_shop
```

## Liste des tests et leur contenu

| Fichier | Tests | Ce qui est testé |
|---------|-------|------------------|
| `protocol_test.py` | 3 | Discovery UCP, version negotiation |
| `checkout_lifecycle_test.py` | 11 | Création, update, completion checkout |
| `validation_test.py` | 6 | Validation données, erreurs |
| `business_logic_test.py` | 8 | Discounts, shipping, totals |
| `fulfillment_test.py` | 11 | Options shipping, hiérarchie |
| `order_test.py` | 4 | Création ordre, statuts |
| `idempotency_test.py` | 4 | Clés idempotence |
| `webhook_test.py` | 3 | Webhooks order_placed/shipped |
| `simulation_url_security_test.py` | 3 | Sécurité endpoint simulation |
| `binding_test.py` | 1 | Binding MCP |
| `invalid_input_test.py` | 3 | Gestion erreurs |
| `card_credential_test.py` | 1 | Credentials paiement |
| `ap2_test.py` | 1 | Apple Pay 2 |

## Interpréter les résultats

**Succès** :
```
Running protocol_test.py...
...
----------------------------------------------------------------------
Ran 3 tests in 0.234s

OK
```

**Échec** :
```
FAIL: test_checkout_create (checkout_lifecycle_test.CheckoutLifecycleTest)
----------------------------------------------------------------------
AssertionError: Expected status 'incomplete', got 'pending'
```

## Débugger un test qui échoue

### 1. Activer les logs détaillés

Ajouter `-v` au test :
```bash
python3 checkout_lifecycle_test.py -v \
  --server_url=http://localhost:8182 \
  --simulation_secret=super-secret-sim-key \
  --conformance_input=test_data/flower_shop/conformance_input.json \
  --test_data_dir=test_data/flower_shop
```

### 2. Vérifier les logs du serveur

Dans le terminal 1, observez les requêtes entrantes et les erreurs.

### 3. Tester manuellement l'endpoint

```bash
# Exemple pour checkout_create
curl -X POST http://localhost:8182/shopping-api/checkout-sessions \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"item_id": "rose_bouquet", "quantity": 1}]
  }' | jq
```

### 4. Comparer avec la spec UCP

Consultez https://ucp.dev pour la spec officielle de la version 2026-01-11.

## Exécuter en CI/CD

### GitHub Actions

Voir `.github/workflows/conformance.yml` :

```yaml
name: UCP Conformance Tests
on: [push, pull_request]
jobs:
  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - name: Start server
        run: |
          go run ./sample_implementation \
            --port 8182 \
            --data-dir test_data/flower_shop \
            --simulation-secret test-secret &
          sleep 2
      - name: Run tests
        run: |
          cd conformance
          for test in *_test.py; do
            python3 $test --server_url=http://localhost:8182 \
              --simulation_secret=test-secret \
              --conformance_input=test_data/flower_shop/conformance_input.json \
              --test_data_dir=test_data/flower_shop
          done
```

## Cas particuliers

### Test idempotency échoue

Vérifiez que le serveur persiste les clés d'idempotence entre requêtes.

### Test webhook échoue

Le serveur doit faire des requêtes HTTP POST vers l'URL de webhook. Vérifiez les logs réseau.

### Test simulation_url_security échoue

Vérifiez que le header `Simulation-Secret` est requis et validé.

## Tests Go internes

En plus des tests UCP, exécuter les tests Go :

```bash
cd ~/stageocto/ucp-merchant-test
go test -v -count=1 ./...
```

Le flag `-count=1` désactive le cache pour forcer la ré-exécution.

## Résolution de problèmes courants

**"Connection refused"** :
- Le serveur n'est pas lancé sur le port 8182
- Vérifiez : `lsof -i :8182`

**"FileNotFoundError: conformance_input.json"** :
- Mauvais chemin vers test_data
- Vérifiez : `ls test_data/flower_shop/conformance_input.json`

**"Invalid UCP version"** :
- Le serveur retourne une mauvaise version
- Doit être exactement `"2026-01-11"` partout

**"Totals mismatch"** :
- Erreur de calcul dans pricing
- Vérifiez `pkg/merchant/pricing/totals.go`

## Validation finale

Tous les tests doivent passer :

```bash
./run_all_tests.sh

# Résultat attendu :
# ✅ protocol_test.py: 3/3
# ✅ checkout_lifecycle_test.py: 11/11
# ✅ validation_test.py: 6/6
# ...
# Total: 60/60 tests passed
```

## Ressources

- [UCP Specification](https://ucp.dev)
- [Conformance repo](https://github.com/universal-commerce-protocol/conformance)
- [Troubleshooting](../reference/error-codes.md)
