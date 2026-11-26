package store

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Limites : une fausse base en mémoire ne doit pas pouvoir grossir sans fin.
const (
	MaxPseudoLength  = 30
	MaxPerTournament = 100
)

var (
	ErrInvalidPseudo = errors.New("pseudo vide ou trop long")
	ErrDuplicate     = errors.New("ce pseudo est déjà inscrit")
	ErrFull          = errors.New("ce tournoi est complet")
)

// registrations : numéro du tournoi -> pseudos inscrits, dans l'ordre d'arrivée.
// Elle est protégée par le même mutex que les likes.
var registrations = map[int][]string{}

// Register inscrit un pseudo à un tournoi.
// Les espaces autour sont retirés ; « Léa » et « léa » comptent pour le même pseudo.
func Register(tournamentID int, pseudo string) error {
	pseudo = strings.TrimSpace(pseudo)
	if n := utf8.RuneCountInString(pseudo); n == 0 || n > MaxPseudoLength {
		return ErrInvalidPseudo
	}

	mu.Lock()
	defer mu.Unlock()
	for _, p := range registrations[tournamentID] {
		if strings.EqualFold(p, pseudo) {
			return ErrDuplicate
		}
	}
	if len(registrations[tournamentID]) >= MaxPerTournament {
		return ErrFull
	}
	registrations[tournamentID] = append(registrations[tournamentID], pseudo)
	return nil
}

// Players renvoie les pseudos inscrits à un tournoi (une copie).
func Players(tournamentID int) []string {
	mu.Lock()
	defer mu.Unlock()
	return append([]string{}, registrations[tournamentID]...)
}
