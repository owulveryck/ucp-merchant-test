# Arrêter les agents

## Méthode simple (1 commande)

Dans le terminal, tapez :

```bash
./scripts/stop-agents.sh
```

**Vous verrez** :
```
✅ Stopping all agents...
✅ Customer Growth Agent stopped
✅ Competitiveness Agent stopped
✅ Dashboard stopped
```

---

## Vérifier que tout est arrêté

Si vous allez sur http://localhost:8080, vous devriez voir :

```
Cette page n'est pas accessible
```

C'est normal ! Les agents sont bien arrêtés.

---

## Relancer les agents

Pour redémarrer tout :

```bash
./scripts/start-agents.sh
```

---

## En cas de problème

### Les agents ne s'arrêtent pas ?

**Solution 1** : Relancez la commande
```bash
./scripts/stop-agents.sh
```

**Solution 2** : Fermez le terminal
- Sur Mac : `Cmd + Q`
- Sur Windows : Cliquez sur la croix

Les agents s'arrêteront automatiquement.

---

## Prochaine étape

[Retour au tutorial](../tutorial/premier-lancement.md)
