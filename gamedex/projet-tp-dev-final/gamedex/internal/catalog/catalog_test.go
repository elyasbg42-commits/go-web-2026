package catalog

import (
	"reflect"
	"testing"

	"gamedex/internal/api"
)

var jeux = []api.Game{
	{ID: 1, Name: "Celeste", Year: 2018, Genres: []string{"plateforme", "indé"}},
	{ID: 3, Name: "Rocket League", Year: 2015, Genres: []string{"sport", "course"}},
	{ID: 9, Name: "Trackmania", Year: 2020, Genres: []string{"course"}},
}

func TestFilter(t *testing.T) {
	cas := []struct {
		nom, genre, query string
		attendu           []int // identifiants attendus
	}{
		{"sans filtre", "", "", []int{1, 3, 9}},
		{"par genre", "course", "", []int{3, 9}},
		{"par nom, sans tenir compte des majuscules", "", "CEL", []int{1}},
		{"nom et genre", "course", "track", []int{9}},
		{"espaces autour ignorés", "", "  cel ", []int{1}},
		{"aucun résultat", "course", "celeste", []int{}},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			ids := []int{}
			for _, g := range Filter(jeux, c.genre, c.query) {
				ids = append(ids, g.ID)
			}
			if !reflect.DeepEqual(ids, c.attendu) {
				t.Errorf("Filter(%q, %q) = %v, attendu %v", c.genre, c.query, ids, c.attendu)
			}
		})
	}
}

func TestFilterNeRendJamaisNil(t *testing.T) {
	// Un nil serait encodé en JSON null au lieu de [].
	if Filter(jeux, "inconnu", "") == nil {
		t.Error("Filter doit renvoyer une slice vide, pas nil")
	}
}

func TestFindGame(t *testing.T) {
	if g, ok := FindGame(jeux, 3); !ok || g.Name != "Rocket League" {
		t.Errorf("FindGame(3) = %v, %v", g, ok)
	}
	if _, ok := FindGame(jeux, 999); ok {
		t.Error("FindGame(999) ne doit rien trouver")
	}
}

func TestGroupByGame(t *testing.T) {
	tournois := []api.Tournament{{ID: 1, GameID: 3}, {ID: 2, GameID: 3}, {ID: 3, GameID: 4}}
	byGame := GroupByGame(tournois)
	if len(byGame[3]) != 2 || len(byGame[4]) != 1 || len(byGame[7]) != 0 {
		t.Errorf("GroupByGame = %v", byGame)
	}
}

func TestGenres(t *testing.T) {
	attendu := []string{"course", "indé", "plateforme", "sport"}
	if got := Genres(jeux); !reflect.DeepEqual(got, attendu) {
		t.Errorf("Genres = %v, attendu %v", got, attendu)
	}
}
