// Package catalog contient les règles : chercher, filtrer, relier.
// Pas de HTTP, pas de HTML, pas d'adresse d'API : juste des listes.
package catalog

import (
	"slices"
	"strings"

	"gamedex/internal/api"
)

// Filter garde les jeux qui ont le genre demandé et dont le nom contient query.
// Un filtre vide est ignoré. La recherche ignore les majuscules : « CEL » trouve « Celeste ».
func Filter(games []api.Game, genre, query string) []api.Game {
	query = strings.ToLower(strings.TrimSpace(query))

	result := []api.Game{}
	for _, g := range games {
		if genre != "" && !slices.Contains(g.Genres, genre) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(g.Name), query) {
			continue
		}
		result = append(result, g)
	}
	return result
}

// FindGame cherche un jeu par son identifiant.
func FindGame(games []api.Game, id int) (api.Game, bool) {
	for _, g := range games {
		if g.ID == id {
			return g, true
		}
	}
	return api.Game{}, false
}

// GroupByGame range les tournois par identifiant de jeu, en un seul passage.
func GroupByGame(tournaments []api.Tournament) map[int][]api.Tournament {
	byGame := make(map[int][]api.Tournament)
	for _, t := range tournaments {
		byGame[t.GameID] = append(byGame[t.GameID], t)
	}
	return byGame
}

// Genres renvoie la liste triée, sans doublon, de tous les genres : pour le menu déroulant.
func Genres(games []api.Game) []string {
	genres := []string{}
	for _, g := range games {
		for _, genre := range g.Genres {
			if !slices.Contains(genres, genre) {
				genres = append(genres, genre)
			}
		}
	}
	slices.Sort(genres)
	return genres
}
