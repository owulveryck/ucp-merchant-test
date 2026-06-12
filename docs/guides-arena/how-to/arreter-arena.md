# Arrêter l'Arena

## Méthode simple (1 commande)

Dans le terminal, tapez :

```bash
./scripts/stop-demo.sh
```

**Vous verrez** :
```
✅ Stopping Arena...
✅ Arena stopped
✅ Observability Hub stopped
✅ Shopping Graph stopped
```

---

## Vérifier que tout est arrêté

```bash
ps aux | grep -E "arena|shopping-graph|obs-hub"
```

**Si tout est arrêté** : Aucune ligne ne s'affiche (ou juste la commande grep elle-même)

**Si des processus tournent encore** : Relancez `./scripts/stop-demo.sh`

---

## Relancer l'Arena

Pour redémarrer tout :

```bash
./scripts/start-demo.sh
```

---

## En cas de problème

### Les agents ne s'arrêtent pas ?

**Solution 1** : Relancez la commande
```bash
./scripts/stop-demo.sh
```

**Solution 2** : Arrêt forcé
```bash
killall arena shopping-graph obs-hub
```

**Solution 3** : Fermez le terminal
- Sur Mac : `Cmd + Q`
- Sur Windows : Cliquez sur la croix

Les agents s'arrêteront automatiquement.

---

## Nettoyer les logs (optionnel)

Si vous voulez supprimer l'historique des logs :

```bash
rm logs/*.log
```

**Attention** : Vous perdrez tout l'historique des décisions !

---

## Prochaine étape

[Retour au tutorial](../tutorial/premier-lancement.md)
