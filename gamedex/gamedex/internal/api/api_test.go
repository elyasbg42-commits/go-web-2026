package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveur démarre une fausse API et pointe BaseURL dessus.
func serveur(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	s := httptest.NewServer(h)
	old := BaseURL
	BaseURL = s.URL
	t.Cleanup(func() { s.Close(); BaseURL = old })
}

func TestFetchGamesOK(t *testing.T) {
	serveur(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1,"name":"Celeste","year":2018,"studio":"MMG","genres":["indé"]}]`))
	})
	games, err := FetchGames()
	if err != nil || len(games) != 1 || games[0].Name != "Celeste" {
		t.Fatalf("FetchGames = %v, %v", games, err)
	}
}

func TestFetchGamesCodeNon200(t *testing.T) {
	// Une API qui répond 500 ne fait pas d'erreur côté Go : c'est à nous de vérifier le code.
	serveur(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"panne"}`, http.StatusInternalServerError)
	})
	_, err := FetchGames()
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("erreur attendue mentionnant 500, reçu : %v", err)
	}
}

func TestFetchGamesJSONIllisible(t *testing.T) {
	serveur(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pas du json")) })
	if _, err := FetchGames(); err == nil {
		t.Fatal("erreur attendue pour un JSON illisible")
	}
}

func TestSearchGamesEncodeLesParametres(t *testing.T) {
	var recu string
	serveur(t, func(w http.ResponseWriter, r *http.Request) {
		recu = r.URL.RawQuery
		w.Write([]byte(`[]`))
	})
	if _, err := SearchGames("ce", "indé"); err != nil {
		t.Fatal(err)
	}
	if recu != "genre=ind%C3%A9&q=ce" {
		t.Errorf("query reçue = %q", recu)
	}
}
