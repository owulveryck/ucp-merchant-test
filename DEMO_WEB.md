# Deploiement web de la demo Arena

Ce guide decrit comment deployer la demo Arena sur AWS pour la rendre accessible sur internet.

## Architecture

```
  Poste local                              AWS EC2
 +-----------------+                      +-----------------------------+
 |                 |                      |                             |
 |  Client Agent   |  HTTPS              |  Caddy (reverse proxy)      |
 |  (Gemini/Vertex)|<-------------------->|   :443 / :80               |
 |                 |                      |     |                       |
 +-----------------+                      |     +-- demo.domain.com     |
                                          |     |     -> Arena :8888    |
  Navigateurs                             |     +-- obs.domain.com     |
 +-----------------+                      |     |     -> Obs Hub :9002  |
 | Marchands       |  HTTPS              |     +-- graph.domain.com   |
 | (dashboards)    |<-------------------->|           -> Graph :9000    |
 |                 |                      |                             |
 | Presentateur    |  HTTPS              |  Inter-services: localhost  |
 | (obs hub)       |<-------------------->|  Arena <-> Graph <-> Obs   |
 +-----------------+                      +-----------------------------+
```

**Principe** : Les 3 services serveur (Arena, Shopping Graph, Obs Hub) tournent sur une seule machine EC2. Caddy gere le HTTPS automatique (Let's Encrypt) et route les sous-domaines. Le Client Agent tourne en local car il necessite l'acces a Google Vertex AI.

## Prerequis

- **AWS** : Un compte AWS avec acces EC2, Route53, et une paire de cles SSH
- **Domaine** : Un nom de domaine gere dans Route53 (ex: `owulveryck.info`)
- **Terraform** : Installe localement (`brew install terraform`)
- **Go** : >= 1.24 pour la compilation
- **SSH** : Cle privee correspondant a la key pair AWS

## Etape 1 : Infrastructure Terraform

Les fichiers Terraform sont dans `demo/infra/` (exclu de git).

### Configuration

Exporter les variables d'environnement :

```bash
export TF_VAR_domain="demo.owulveryck.info"
export TF_VAR_zone_id="ZXXXXXXXXXX"         # ID de la zone Route53 pour owulveryck.info
export TF_VAR_ssh_key_name="my-aws-key"     # Nom de la key pair EC2
export TF_VAR_region="eu-west-3"            # Paris (optionnel, defaut eu-west-3)
export TF_VAR_instance_type="t3.small"      # Optionnel, defaut t3.small
```

### Lancement

```bash
cd demo/infra
terraform init
terraform plan     # Verifier les ressources
terraform apply    # Creer l'infra (confirmer avec 'yes')
```

Terraform cree :
- 1 instance EC2 Ubuntu 24.04 avec Caddy pre-installe
- 1 Elastic IP (IP stable)
- 1 Security Group (ports 22, 80, 443)
- 2 enregistrements DNS Route53 : `demo.owulveryck.info` + `*.demo.owulveryck.info`

Noter l'IP publique affichee en sortie :
```
Outputs:
  public_ip   = "X.X.X.X"
  arena_url   = "https://demo.owulveryck.info"
  ssh_command = "ssh -i ~/.ssh/my-aws-key.pem ubuntu@X.X.X.X"
```

## Etape 2 : Deploiement des services

### Variables d'environnement

```bash
export DEPLOY_HOST="X.X.X.X"                    # IP de l'EC2 (depuis terraform output)
export DEPLOY_KEY="$HOME/.ssh/my-aws-key.pem"    # Cle SSH privee
export DEPLOY_DOMAIN="demo.owulveryck.info"      # Domaine public
```

### Lancement

```bash
./demo/scripts/run_arena_web.sh
```

Le script :
1. Cross-compile les binaires Go pour linux/amd64
2. Genere le Caddyfile a partir du template
3. Upload binaires + Caddyfile sur l'EC2 via SCP
4. Arrete les anciens services, demarre les nouveaux
5. Recharge la configuration Caddy

### Verification

Le script affiche les URLs a la fin :
```
  Arena:          https://demo.owulveryck.info/
  Arena (auto):   https://demo.owulveryck.info/auto
  Obs Hub:        https://obs.demo.owulveryck.info/arena
  Shopping Graph: https://graph.demo.owulveryck.info/health
```

Tester :
```bash
# Health check du shopping graph
curl https://graph.demo.owulveryck.info/health

# Page d'accueil de l'arena (devrait retourner du HTML)
curl -s https://demo.owulveryck.info/ | head -5
```

## Etape 3 : Client Agent (local)

Le client agent tourne sur votre poste local et se connecte aux services distants :

```bash
# Compiler le client (si pas deja fait)
go build -o demo/bin/client ./demo/cmd/client/

# Lancer le client
demo/bin/client \
  --graph-url https://graph.demo.owulveryck.info \
  --obs-url https://obs.demo.owulveryck.info
```

Le client necessite `GOOGLE_CLOUD_PROJECT` et des credentials Vertex AI configures localement.

## Mode local (inchange)

Le script `run_arena.sh` continue de fonctionner exactement comme avant :

```bash
./demo/scripts/run_arena.sh
```

Tous les services tournent en local sur `localhost`, aucun changement de comportement.

## Operations

### Consulter les logs distants

```bash
ssh -i $DEPLOY_KEY ubuntu@$DEPLOY_HOST 'tail -f /opt/demo/*.log'
```

### Arreter les services distants

```bash
ssh -i $DEPLOY_KEY ubuntu@$DEPLOY_HOST 'pkill -f /opt/demo/'
```

### Redeployer (apres modification du code)

Relancer simplement :
```bash
./demo/scripts/run_arena_web.sh
```

Le script arrete les anciens processus et redeploit les nouveaux binaires.

### Detruire l'infrastructure

```bash
cd demo/infra
terraform destroy
```

## Flag `--base-url`

Le flag `--base-url` ajoute a l'Arena permet de specifier l'URL publique du serveur. Quand il est set :
- L'**agent card A2A** expose l'URL publique (ex: `https://demo.owulveryck.info/{tenant}/a2a`)
- Les **metadonnees OAuth** utilisent l'URL publique pour les endpoints
- Le **discovery UCP** renvoie l'URL publique

Quand il n'est pas set (defaut), tout utilise `http://localhost:PORT` comme avant.

Les communications internes entre services (Graph polling les marchands, Arena -> Graph, etc.) restent en `localhost` car tous les services sont sur la meme machine.

## Sous-domaines

| Sous-domaine | Service | Port | Usage |
|---|---|---|---|
| `demo.owulveryck.info` | Arena | 8888 | Inscription marchands, dashboards, A2A |
| `obs.demo.owulveryck.info` | Obs Hub | 9002 | Dashboard presentateur, SSE events |
| `graph.demo.owulveryck.info` | Shopping Graph | 9000 | API recherche, polling, health |

## Troubleshooting

### Caddy ne demarre pas / certificat SSL echoue
- Verifier que les ports 80 et 443 sont ouverts (Security Group)
- Verifier que le DNS pointe bien vers l'IP de l'EC2 : `dig demo.owulveryck.info`
- Consulter les logs Caddy : `ssh ... 'sudo journalctl -u caddy -f'`

### Les services ne repondent pas
- Verifier qu'ils tournent : `ssh ... 'pgrep -la "shopping-graph|obs-hub|arena"'`
- Consulter les logs : `ssh ... 'cat /opt/demo/arena.log'`

### Le client agent ne se connecte pas
- Verifier la resolution DNS : `curl https://graph.demo.owulveryck.info/health`
- Verifier les credentials Vertex AI : `gcloud auth application-default print-access-token`

### L'agent card retourne localhost
- Verifier que l'arena a ete lancee avec `--base-url` : `ssh ... 'cat /opt/demo/arena.log | head -5'`
