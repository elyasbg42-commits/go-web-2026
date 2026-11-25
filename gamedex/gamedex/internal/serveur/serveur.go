// Package server contient les routes et les handlers :
// lire la requête, appeler le cœur (catalog), afficher.
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
)

// Server garde les données chargées au démarrage : les pages se servent en mémoire.
type Server struct {
	Games       []api.Game
	Tournaments []api.Tournament

	tmpl   *template.Template
	byGame map[int][]api.Tournament
}

// Page : ce que tous les templates partagent (le header en a besoin).
type Page struct {
	Title string
	Map   bool // true : le header charge le CSS de Leaflet
}

type ListPage struct {
	Page
	Games  []api.Game
	Genres []string
	Query  string
	Genre  string
	Source string // "" ou "API" : d'où viennent les résultats
}

type GamePage struct {
	Page
	Game        api.Game
	Tournaments []api.Tournament
}

type ErrorPage struct {
	Page
	Code    int
	Message string
}

// Routes lit les templates une seule fois et branche les handlers.
// À appeler depuis le dossier gamedex/ (les chemins web/... sont relatifs).
func (s *Server) Routes() http.Handler {
	s.tmpl = template.Must(template.ParseGlob("web/templates/*.html"))
	s.byGame = catalog.GroupByGame(s.Tournaments)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /jeux", s.list)
	mux.HandleFunc("GET /jeux/{id}", s.detail)
	mux.HandleFunc("POST /recherche", s.search)
	// Un GET sur /recherche : la route existe, mais seulement en POST. Page 405 sur mesure.
	mux.HandleFunc("GET /recherche", s.methodNotAllowed)
	mux.HandleFunc("GET /api/jeux", s.apiList)
	// "GET /" attrape tout le reste : c'est notre page 404.
	mux.HandleFunc("GET /", s.notFound)

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	return mux
}

// show remplit un template et envoie la page avec le bon code.
// On écrit d'abord dans un buffer : si le template échoue, on peut encore
// répondre 500 proprement au lieu d'envoyer une page à moitié construite.
func (s *Server) show(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Println("template", name, ":", err)
		http.Error(w, "erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) showError(w http.ResponseWriter, status int, message string) {
	s.show(w, status, "erreur.html", ErrorPage{
		Page:    Page{Title: fmt.Sprintf("Erreur %d", status)},
		Code:    status,
		Message: message,
	})
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.show(w, http.StatusOK, "accueil.html", Page{Title: "GameDex"})
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", http.MethodPost)
	s.showError(w, http.StatusMethodNotAllowed, "Cette adresse n'accepte que l'envoi du formulaire de recherche (POST).")
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.showError(w, http.StatusNotFound, "Cette page n'existe pas.")
}

// list : /jeux?q=cel&genre=course (les deux filtres sont facultatifs).
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")

	s.show(w, http.StatusOK, "liste.html", ListPage{
		Page:   Page{Title: "Les jeux"},
		Games:  catalog.Filter(s.Games, genre, q),
		Genres: catalog.Genres(s.Games),
		Query:  q,
		Genre:  genre,
	})
}

// detail : /jeux/{id}. 400 si l'id n'est pas un nombre, 404 si le jeu n'existe pas.
func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		s.showError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre positif.")
		return
	}
	game, ok := catalog.FindGame(s.Games, id)
	if !ok {
		s.showError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}
	tournaments := s.byGame[id]
	s.show(w, http.StatusOK, "jeu.html", GamePage{
		// La carte n'a de sens que s'il y a des tournois à placer.
		Page:        Page{Title: game.Name, Map: len(tournaments) > 0},
		Game:        game,
		Tournaments: tournaments,
	})
}

// search (bonus) : le formulaire POST /recherche est transmis à la mini API,
// c'est elle qui filtre. Si elle ne répond pas : 502 Bad Gateway.
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	genre := r.FormValue("genre")

	games, err := api.SearchGames(q, genre)
	if err != nil {
		log.Println("recherche via l'API :", err)
		s.showError(w, http.StatusBadGateway, "Le service de recherche ne répond pas. Réessayez dans un instant.")
		return
	}
	s.show(w, http.StatusOK, "liste.html", ListPage{
		Page:   Page{Title: "Résultats de la recherche"},
		Games:  games,
		Genres: catalog.Genres(s.Games),
		Query:  q,
		Genre:  genre,
		Source: "API",
	})
}

// apiList (bonus) : notre serveur parle aussi JSON, avec le même catalog.Filter.
func (s *Server) apiList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")

	games := catalog.Filter(s.Games, genre, q)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(games); err != nil {
		log.Println("encodage JSON :", err)
	}
}
