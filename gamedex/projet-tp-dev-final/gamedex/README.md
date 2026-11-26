# GameDex

Catalogue de jeux vidéo et de leurs tournois, écrit en Go avec la bibliothèque standard (`net/http` et `html/template`), sans framework. Les données viennent de la mini API du cours.

Projet réalisé dans le cadre du cours « Application web Go », B1 Ynov Lyon.

## Fonctionnalités

- **Liste des jeux** avec recherche par nom et filtre par genre (formulaire GET, l'adresse est partageable : `/jeux?q=cel&genre=course`). Les jeux les plus likés sont affichés en premier.
- **Fiche d'un jeu** : année, studio, genres, nombre de likes, liste de ses tournois et carte Leaflet de leurs lieux.
- **Bouton J'aime** (formulaire POST) : le like est enregistré en mémoire, puis le navigateur est redirigé vers la fiche. Recharger la page (F5) n'ajoute donc pas de like.
- **Inscription à un tournoi** avec un pseudo (formulaire POST), enregistrée en mémoire, avec la liste des inscrits.
- **Recherche instantanée** : à chaque frappe, le JavaScript demande du JSON à GameDex (`/api/jeux`) et met la liste à jour sans recharger la page. Sans JavaScript, le formulaire GET fonctionne toujours.
- **Recherche via la mini API** (formulaire POST `/recherche`) : GameDex transmet la recherche à la mini API, qui filtre.
- **Pages d'erreur** avec le bon code HTTP : 400, 404, 405, 409, 502.

## Prérequis

- Go 1.22 ou plus (les routes `GET /jeux/{id}` demandent Go 1.22)
- Le dépôt du cours, qui contient la mini API dans le dossier `api/`
- Un accès internet dans le navigateur pour la carte (Leaflet et les tuiles OpenStreetMap sont chargés depuis internet)

## Lancer

Il faut deux terminaux.

```
# Terminal 1, à la racine du dépôt : la mini API (port 8081)
go run ./api

# Terminal 2, dans le dossier gamedex/ : GameDex (port 8080)
cd gamedex
go run ./cmd/gamedex
```

Puis ouvrir http://localhost:8080.

GameDex doit être lancé **depuis le dossier `gamedex/`** : les chemins `web/templates` et `web/static` sont relatifs à ce dossier.

Options :

| Option | Défaut | Rôle |
| --- | --- | --- |
| `-port` | `8080` | port d'écoute de GameDex |
| `-api` | `http://localhost:8081` | adresse de la mini API |

Exemple : `go run ./cmd/gamedex -port 9000`.

GameDex charge les jeux et les tournois **une seule fois au démarrage** (deux appels à la mini API). Si elle ne répond pas à ce moment-là, GameDex s'arrête avec un message clair, par exemple `Impossible de charger les jeux : appel de /api/games : … connection refused`.

## Routes

| Route | Rôle |
| --- | --- |
| `GET /` | Accueil |
| `GET /jeux` | Liste des jeux, les plus likés en premier. Filtres facultatifs : `?q=cel&genre=course`, transmis à la mini API |
| `GET /jeux/{id}` | Fiche d'un jeu, ses tournois, le formulaire J'aime et la carte |
| `POST /jeux/{id}/like` | Ajoute un like, puis redirige (303) vers la fiche |
| `POST /tournois/{id}/inscription` | Inscrit le `pseudo` du formulaire à un tournoi, puis redirige (303) vers la fiche du jeu |
| `POST /recherche` | Transmet `q` et `genre` à la mini API et affiche les résultats |
| `GET /api/jeux?q=&genre=` | La liste filtrée, avec les likes, en JSON (filtrée en mémoire, la mini API n'est pas appelée) |
| `GET /static/...` | CSS et JavaScript |
| toute autre adresse | Page 404 |

### Codes de réponse

| Code | Quand |
| --- | --- |
| 200 | Page affichée |
| 303 | Like ou inscription enregistré : redirection vers la fiche |
| 400 | Identifiant qui n'est pas un nombre positif, ou pseudo vide ou de plus de 30 caractères |
| 404 | Jeu, tournoi ou page inexistant |
| 405 | `GET /recherche` (cette adresse n'accepte que POST) |
| 409 | Pseudo déjà inscrit à ce tournoi, ou tournoi complet (100 inscrits) |
| 502 | La mini API ne répond pas (liste `/jeux` et `POST /recherche`) |

## Organisation du code

```
gamedex/
├── go.mod                       module gamedex
├── cmd/gamedex/main.go          charge les données, lance le serveur
├── internal/
│   ├── api/                     parle à la mini API (client avec timeout de 3 s)
│   │   ├── api.go
│   │   └── api_test.go
│   ├── catalog/                 les règles : Filter, FindGame, GroupByGame, Genres
│   │   ├── catalog.go
│   │   └── catalog_test.go
│   ├── store/                   la fausse base de données, en mémoire
│   │   ├── likes.go             AddLike, Likes, AllLikes
│   │   ├── registrations.go     Register, Players
│   │   └── store_test.go
│   └── server/
│       └── server.go            routes et handlers
└── web/
    ├── templates/pages.html     toutes les pages (un bloc {{define}} par page)
    └── static/
        ├── css/style.css
        └── js/
            ├── filtre.js        recherche instantanée (fetch)
            └── carte.js         carte Leaflet des tournois
```

Qui connaît qui :

| Package | Rôle | Ne connaît pas |
| --- | --- | --- |
| `api` | Appeler la mini API, vérifier le code HTTP, décoder le JSON | le HTML, le navigateur |
| `catalog` | Chercher, filtrer, relier les tournois aux jeux | HTTP, HTML, l'adresse de l'API |
| `store` | Compter les likes, garder les inscriptions | HTTP, HTML, l'API |
| `server` | Lire la requête, appeler les autres packages, afficher une page | les règles de filtrage |
| `web/` | Présentation | aucune règle |

## Choix techniques

- **Un client HTTP avec timeout** (3 secondes) : une API lente ne doit pas bloquer GameDex.
- **Trois vérifications à chaque appel d'API** : la réponse est arrivée, le code est 200, le JSON est lisible. Une API qui répond 500 ne produit aucune erreur côté Go : c'est la vérification du code qui l'attrape.
- **`url.Values` pour construire les adresses** : les accents et les espaces sont encodés, on ne colle jamais du texte tapé directement dans une URL.
- **Les pages sont écrites dans un buffer avant d'être envoyées** : si un template échoue, GameDex peut encore répondre 500 proprement au lieu d'envoyer une page à moitié construite.
- **`html/template` échappe tout** : un `<script>` tapé dans une recherche ou un pseudo s'affiche comme du texte. Côté JavaScript, le texte venant du serveur est inséré avec `textContent`, jamais en HTML brut.
- **POST puis redirection 303** (motif « Post/Redirect/Get ») : après un like ou une inscription, le navigateur finit sur un GET qu'on peut recharger sans refaire l'action.
- **Un mutex dans `store`** : le serveur traite plusieurs requêtes en même temps, et le mutex garantit qu'une seule à la fois touche la map.
- **La liste `/jeux` passe par la mini API, la recherche instantanée passe par la mémoire** : la première suit le cours (transmettre `q` et `genre` à l'API), la seconde doit répondre à chaque frappe sans appel réseau.

## Vérifier

```
cd gamedex
gofmt -l .
go vet ./...
go test -race ./...
```

`gofmt -l .` et `go vet ./...` ne doivent rien afficher. Les tests n'ont pas besoin de la mini API : `api` est testé contre une fausse API (`httptest`).

Test des routes à la main, une fois les deux serveurs lancés :

```
for u in "" jeux "jeux?q=cel" "jeux?genre=course" jeux/3 jeux/abc jeux/999 nimporte/quoi "api/jeux?q=hol"; do
  curl -s -o /dev/null -w "/$u -> %{http_code}\n" "localhost:8080/$u"
done
curl -s -o /dev/null -w "POST like -> %{http_code}\n" -X POST localhost:8080/jeux/3/like
curl -s -o /dev/null -w "GET /recherche -> %{http_code}\n" localhost:8080/recherche
```

Résultats attendus : 200 pour les pages et l'API JSON, 400 pour `jeux/abc`, 404 pour `jeux/999` et `nimporte/quoi`, 303 pour le like, 405 pour `GET /recherche`.

Pour voir le 502, arrêter la mini API (Ctrl+C dans le terminal 1) puis ouvrir `/jeux`.

## Limites

- Les **likes et les inscriptions sont en mémoire** : ils disparaissent quand GameDex s'arrête. Une vraie base de données (SQLite, PostgreSQL) ferait la même chose, mais sur le disque.
- Les jeux et les tournois sont **chargés une seule fois au démarrage** : un changement dans la mini API n'apparaît qu'après un redémarrage de GameDex.
- Il n'y a **ni compte utilisateur ni protection CSRF** : un même visiteur peut liker un jeu autant de fois qu'il le veut. C'est suffisant pour un exercice, pas pour un vrai site.
- La carte demande **une connexion internet** dans le navigateur.

## Parcours du cours

| Mini TP | Contenu |
| --- | --- |
| 1 | Explorer la mini API (Swagger, codes 400, 401, 403, 404, 500) |
| 2 | Premier serveur : routes, 400 et 404, fichiers statiques, option `-port` |
| 3 | Appeler la mini API : structs, timeout, gestion des erreurs |
| 4 | Pages avec `html/template` et pages d'erreur |
| 5 | Ranger le code : `cmd/`, `internal/`, `web/` |
| 6 | Filtrer (GET) et liker (POST), bonus : tri par likes, inscription à un tournoi |
| Bonus | Carte Leaflet, recherche POST via l'API, recherche instantanée avec `fetch` |
