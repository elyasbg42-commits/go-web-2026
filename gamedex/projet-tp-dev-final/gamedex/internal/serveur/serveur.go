// Package server contient les routes et les handlers :
// lire la requête, appeler le cœur (catalog, store, api), afficher.
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/store"
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

// ListItem : un jeu avec son nombre de likes. Le JSON de /api/jeux a la même forme.
type ListItem struct {
	api.Game
	Likes int `json:"likes"`
}

type ListPage struct {
	Page
	Games  []ListItem
	Genres []string
	Query  string
	Genre  string
	Source string // "" ou "API" : d'où viennent les résultats
}

// TournamentEntry : un tournoi avec les pseudos déjà inscrits.
type TournamentEntry struct {
	api.Tournament
	Players []string
}

type GamePage struct {
	Page
	Game        api.Game
	Likes       int
	Tournaments []api.Tournament // pour la carte (JSON dans le JavaScript)
	Entries     []TournamentEntry
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
	// Mini TP 6 : agir = POST. Le formulaire envoie, le serveur enregistre puis redirige.
	mux.HandleFunc("POST /jeux/{id}/like", s.like)
	mux.HandleFunc("POST /tournois/{id}/inscription", s.register)
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

// idParam lit {id} dans l'URL. Faux si ce n'est pas un nombre positif.
func idParam(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id >= 1
}

// withLikes ajoute les likes à chaque jeu, puis met les plus likés en premier.
// Tri stable : à égalité de likes, l'ordre de départ est conservé.
func withLikes(games []api.Game) []ListItem {
	counts := store.AllLikes()
	items := make([]ListItem, len(games))
	for i, g := range games {
		items[i] = ListItem{Game: g, Likes: counts[g.ID]}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Likes > items[j].Likes })
	return items
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.show(w, http.StatusOK, "accueil.html", Page{Title: "GameDex"})
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.showError(w, http.StatusNotFound, "Cette page n'existe pas.")
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", http.MethodPost)
	s.showError(w, http.StatusMethodNotAllowed, "Cette adresse n'accepte que l'envoi du formulaire de recherche (POST).")
}

// list : /jeux?q=cel&genre=course. Les deux filtres sont facultatifs.
// Mini TP 6 : on les transmet à la mini API, c'est elle qui filtre.
// Un paramètre vide est ignoré par l'API : sans recherche, on reçoit tous les jeux.
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	genre := r.URL.Query().Get("genre")

	games, err := api.SearchGames(q, genre)
	if err != nil {
		log.Println("liste via l'API :", err)
		s.showError(w, http.StatusBadGateway, "Le service des jeux ne répond pas.")
		return
	}
	s.show(w, http.StatusOK, "liste.html", ListPage{
		Page:   Page{Title: "Les jeux"},
		Games:  withLikes(games), // bonus : les plus likés en premier
		Genres: catalog.Genres(s.Games),
		Query:  q,
		Genre:  genre, // le genre choisi reste sélectionné
	})
}

// detail : /jeux/{id}. 400 si l'id n'est pas un nombre, 404 si le jeu n'existe pas.
func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		s.showError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre positif.")
		return
	}
	game, ok := catalog.FindGame(s.Games, id)
	if !ok {
		s.showError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}

	tournaments := s.byGame[id]
	entries := make([]TournamentEntry, len(tournaments))
	for i, t := range tournaments {
		entries[i] = TournamentEntry{Tournament: t, Players: store.Players(t.ID)}
	}
	s.show(w, http.StatusOK, "jeu.html", GamePage{
		// La carte n'a de sens que s'il y a des tournois à placer.
		Page:        Page{Title: game.Name, Map: len(tournaments) > 0},
		Game:        game,
		Likes:       store.Likes(id),
		Tournaments: tournaments,
		Entries:     entries,
	})
}

// like : POST /jeux/{id}/like. On enregistre, puis on REDIRIGE (303 See Other)
// vers la fiche. Sans redirection, un F5 renverrait le formulaire : un like de plus.
// Avec la redirection, le navigateur finit sur un GET, qu'on peut recharger sans risque.
func (s *Server) like(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		s.showError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre positif.")
		return
	}
	if _, ok := catalog.FindGame(s.Games, id); !ok {
		s.showError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
		return
	}
	store.AddLike(id)
	http.Redirect(w, r, fmt.Sprintf("/jeux/%d", id), http.StatusSeeOther)
}

// register (bonus) : POST /tournois/{id}/inscription, avec le champ « pseudo ».
// Même modèle que le like : enregistrer en mémoire, puis rediriger vers la fiche du jeu.
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		s.showError(w, http.StatusBadRequest, "L'identifiant du tournoi doit être un nombre positif.")
		return
	}
	var tournament *api.Tournament
	for i := range s.Tournaments {
		if s.Tournaments[i].ID == id {
			tournament = &s.Tournaments[i]
			break
		}
	}
	if tournament == nil {
		s.showError(w, http.StatusNotFound, "Ce tournoi n'existe pas.")
		return
	}

	switch err := store.Register(id, r.FormValue("pseudo")); {
	case errors.Is(err, store.ErrInvalidPseudo):
		s.showError(w, http.StatusBadRequest,
			fmt.Sprintf("Le pseudo doit faire entre 1 et %d caractères.", store.MaxPseudoLength))
		return
	case errors.Is(err, store.ErrDuplicate), errors.Is(err, store.ErrFull):
		s.showError(w, http.StatusConflict, "Inscription impossible : "+err.Error()+".")
		return
	case err != nil:
		log.Println("inscription :", err)
		s.showError(w, http.StatusInternalServerError, "Une erreur est survenue.")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/jeux/%d", tournament.GameID), http.StatusSeeOther)
}

// search (bonus) : le formulaire POST /recherche est transmis à la mini API,
// c'est elle qui filtre. Si elle ne répond pas : 502 Bad Gateway.
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.FormValue("q"))
	genre := r.FormValue("genre")

	games, err := api.SearchGames(q, genre)
	if err != nil {
		log.Println("recherche via l'API :", err)
		s.showError(w, http.StatusBadGateway, "Le service de recherche ne répond pas. Réessayez dans un instant.")
		return
	}
	s.show(w, http.StatusOK, "liste.html", ListPage{
		Page:   Page{Title: "Résultats de la recherche"},
		Games:  withLikes(games),
		Genres: catalog.Genres(s.Games),
		Query:  q,
		Genre:  genre,
		Source: "API",
	})
}

// apiList (bonus) : notre serveur parle aussi JSON, avec le même tri et les mêmes likes.
// Il filtre en mémoire (catalog.Filter) : c'est instantané, la mini API n'est pas appelée.
func (s *Server) apiList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")

	items := withLikes(catalog.Filter(s.Games, genre, q))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Println("encodage JSON :", err)
	}
}
