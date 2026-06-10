# Reference - Documentation Technique

La référence technique est une **encyclopédie** du projet. Consultez-la pour trouver des informations précises sur l'API, la configuration, les structures de données, etc.

## Documentation Disponible

### API & Protocoles

#### [API Reference Complète](api-reference.md)
**Contenu** :
- REST API (checkout sessions, orders)
- MCP API (JSON-RPC 2.0 tools)
- A2A API (agent-to-agent)
- Shopping Graph API (recherche cross-merchant)
- Observability Hub API (SSE events)
- Codes d'erreur et formats de réponse

**Consultez pour** : Savoir quels endpoints existent, leurs paramètres, et leurs réponses.

### Configuration

#### [Options de Configuration](configuration.md) *(À créer)*
**Contenu** :
- CLI flags (--port, --data-dir, --simulation-secret, etc.)
- Variables d'environnement
- Fichiers de configuration (YAML, JSON)
- Pricing config (min_margin, strategies, etc.)
- Shopping Graph config (merchants, polling_interval)

**Consultez pour** : Connaître toutes les options de configuration disponibles.

### Architecture

#### [Architecture des Agents](agent-architecture.md) *(À créer)*
**Contenu** :
- Diagrammes d'architecture 3-agents
- Interfaces Go de chaque agent
- Flow de décision (séquence diagrams)
- Structures de données (PriceRequest, PriceDecision, etc.)
- Communication inter-agents

**Consultez pour** : Comprendre techniquement comment les agents fonctionnent.

### Data Models

#### [Modèles de Données UCP](data-models.md) *(À créer)*
**Contenu** :
- Checkout session structure
- Order structure
- Totals types (items_discount, subtotal, fulfillment, etc.)
- Payment structure
- Fulfillment hierarchy (methods → destinations → groups → options)
- Item structure

**Consultez pour** : Savoir exactement quels champs existent dans chaque structure.

### Error Handling

#### [Codes d'Erreur](error-codes.md) *(À créer)*
**Contenu** :
- Sentinel errors Go (ErrNotFound, ErrInvalidInput, etc.)
- HTTP status codes mapping
- Error response format
- Common errors et leurs causes
- Debugging tips

**Consultez pour** : Comprendre une erreur spécifique et comment la résoudre.

## Comment Utiliser cette Section

La référence est **consultative**, pas **narrative** :
- ✅ Recherche rapide ("Quel est le format de totals ?")
- ✅ Vérification ("Est-ce que ce champ est requis ?")
- ✅ Copier-coller (exemples de requêtes)
- ❌ Apprentissage de zéro (utilisez [Tutorials](../tutorials/))

**Utilisez Ctrl+F** pour trouver rapidement ce que vous cherchez.

## Format de la Référence

Chaque page de référence suit cette structure :
1. **Vue d'ensemble** (1 paragraphe)
2. **Index des sections** (table of contents)
3. **Sections détaillées** avec :
   - Nom exact (fonction, endpoint, champ)
   - Type / signature
   - Description concise
   - Paramètres / champs
   - Exemple concret
   4. **Voir aussi** (liens vers related docs)

**Style** :
- Concis, précis, factuel
- Pas d'opinions ou d'explications longues
- Exemples réels et testés
- Tables pour les listes

## Différence avec les Autres Sections

| Tutorial | How-to | Reference | Explanation |
|----------|--------|-----------|-------------|
| Apprendre | Faire | Consulter | Comprendre |
| Narratif | Recette | Encyclopédie | Analyse |
| "Voici comment créer un checkout" | "Pour créer un checkout, faites X" | "POST /checkout-sessions - params: items[]" | "Pourquoi checkout avant order" |

## Contribuer à la Référence

La référence doit être :
- **Exacte** : Vérifiez dans le code source
- **Complète** : Tous les champs, tous les paramètres
- **À jour** : Synchronisée avec le code
- **Concise** : Pas d'explications longues
- **Utilisable** : Exemples copiables-collables

**Structure recommandée pour une entrée** :
```markdown
### Nom de la Fonction/Endpoint

**Type/Signature** : `func DoSomething(param string) (Result, error)`

**Description** : Une phrase décrivant ce que ça fait.

**Paramètres** :
- `param` (string, required) - Description du paramètre

**Retourne** :
- `Result` - Description du résultat
- `error` - Erreur si échec

**Exemple** :
\`\`\`go
result, err := DoSomething("value")
\`\`\`

**Voir aussi** : [Related Function](#related)
```

## Maintien de la Référence

La référence doit être mise à jour quand :
- Un nouveau endpoint est ajouté
- Un paramètre change (ajout, suppression, renommage)
- Une structure de données évolue
- Un comportement change

**Process** :
1. Modifier le code
2. Mettre à jour la référence
3. Vérifier que les exemples fonctionnent toujours
4. Commit ensemble (code + doc)

Ne laissez **jamais** la référence diverger du code.

## Automatisation (Futur)

Idéalement, la référence API devrait être générée automatiquement :
- Swagger/OpenAPI pour REST API
- JSON Schema pour structures
- Godoc pour code Go

Pour l'instant, c'est manuel → gardez-la à jour !
