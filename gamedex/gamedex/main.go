// GameDex : charge les données une fois, puis sert les pages.
// À lancer depuis le dossier gamedex/ : go run ./cmd/gamedex
// La mini API doit tourner : go run ./api (à la racine du dépôt).
package main

import (
	"flag"
	"fmt"
	"gamedex/internal/api"
	"gamedex/internal/server"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "port d'écoute de GameDex")
	apiURL := flag.String("api", api.BaseURL, "adresse de la mini API")
	flag.Parse()
	api.BaseURL = *apiURL

	// L'API est appelée au démarrage, deux fois en tout.
	// Si elle ne répond pas : on s'arrête tout de suite avec un message clair.
	games, err := api.FetchGames()
	if err != nil {
		log.Fatal("Impossible de charger les jeux : ", err)
	}
	tournaments, err := api.FetchTournaments()
	if err != nil {
		log.Fatal("Impossible de charger les tournois : ", err)
	}

	srv := &server.Server{Games: games, Tournaments: tournaments}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("GameDex sur http://localhost%s (%d jeux, %d tournois)\n", addr, len(games), len(tournaments))
	log.Fatal(http.ListenAndServe(addr, srv.Routes()))
}
