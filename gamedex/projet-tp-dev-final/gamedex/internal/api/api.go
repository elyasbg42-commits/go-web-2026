// Package api parle à la mini API externe et c'est tout :
// appeler, vérifier le code, décoder le JSON.
// Il ne sait rien du navigateur ni du HTML.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BaseURL est l'adresse de la mini API. main peut la changer (option -api).
var BaseURL = "http://localhost:8081"

// Un client avec timeout : une API lente ne doit pas bloquer GameDex.
var client = &http.Client{Timeout: 3 * time.Second}

// Game a la forme du JSON renvoyé par /api/games.
type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

// Tournament a la forme du JSON renvoyé par /api/tournaments.
type Tournament struct {
	ID     int     `json:"id"`
	GameID int     `json:"gameId"`
	Name   string  `json:"name"`
	City   string  `json:"city"`
	Date   string  `json:"date"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

// get appelle un endpoint et décode la réponse dans out.
// Les trois vérifications : réponse reçue, bon code, JSON lisible.
func get(path string, out any) error {
	resp, err := client.Get(BaseURL + path)
	if err != nil {
		return fmt.Errorf("appel de %s : %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a répondu %d sur %s", resp.StatusCode, path)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("JSON illisible pour %s : %w", path, err)
	}
	return nil
}

// FetchGames renvoie tous les jeux.
func FetchGames() ([]Game, error) {
	var games []Game
	err := get("/api/games", &games)
	return games, err
}

// FetchTournaments renvoie tous les tournois.
func FetchTournaments() ([]Tournament, error) {
	var tournaments []Tournament
	err := get("/api/tournaments", &tournaments)
	return tournaments, err
}

// SearchGames laisse l'API filtrer : /api/games?q=...&genre=...
// url.Values encode les accents et les espaces : on ne colle jamais
// du texte tapé directement dans une URL.
func SearchGames(q, genre string) ([]Game, error) {
	params := url.Values{}
	params.Set("q", q)
	params.Set("genre", genre)

	var games []Game
	err := get("/api/games?"+params.Encode(), &games)
	return games, err
}
