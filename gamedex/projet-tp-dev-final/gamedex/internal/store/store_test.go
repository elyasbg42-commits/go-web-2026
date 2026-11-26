package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// reset remet la fausse base à zéro entre deux tests.
func reset() {
	mu.Lock()
	defer mu.Unlock()
	likes = map[int]int{}
	registrations = map[int][]string{}
}

func TestAddLike(t *testing.T) {
	reset()
	if Likes(3) != 0 {
		t.Fatal("un jeu jamais liké doit avoir 0 like")
	}
	if got := AddLike(3); got != 1 {
		t.Errorf("premier like : %d, attendu 1", got)
	}
	if got := AddLike(3); got != 2 {
		t.Errorf("second like : %d, attendu 2", got)
	}
	if Likes(3) != 2 || Likes(4) != 0 {
		t.Errorf("Likes(3)=%d Likes(4)=%d", Likes(3), Likes(4))
	}
}

// 100 visiteurs likent en même temps : sans mutex, des likes seraient perdus
// (lancer avec `go test -race` pour que Go le signale).
func TestAddLikeEnParallele(t *testing.T) {
	reset()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); AddLike(7) }()
	}
	wg.Wait()
	if Likes(7) != 100 {
		t.Errorf("Likes(7) = %d, attendu 100", Likes(7))
	}
}

func TestAllLikesEstUneCopie(t *testing.T) {
	reset()
	AddLike(1)
	copie := AllLikes()
	copie[1] = 999
	if Likes(1) != 1 {
		t.Error("modifier la copie ne doit pas changer les vrais likes")
	}
}

func TestRegister(t *testing.T) {
	reset()
	if err := Register(1, "  Léa "); err != nil {
		t.Fatal(err)
	}
	if err := Register(1, "léa"); !errors.Is(err, ErrDuplicate) {
		t.Errorf("doublon : %v", err)
	}
	if err := Register(2, "Léa"); err != nil {
		t.Errorf("le même pseudo dans un autre tournoi doit passer : %v", err)
	}
	for _, mauvais := range []string{"", "   ", strings.Repeat("a", MaxPseudoLength+1)} {
		if err := Register(1, mauvais); !errors.Is(err, ErrInvalidPseudo) {
			t.Errorf("Register(%q) = %v", mauvais, err)
		}
	}
	if got := Players(1); len(got) != 1 || got[0] != "Léa" {
		t.Errorf("Players(1) = %v", got)
	}
}

func TestRegisterTournoiComplet(t *testing.T) {
	reset()
	for i := 0; i < MaxPerTournament; i++ {
		if err := Register(1, fmt.Sprintf("joueur%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := Register(1, "dernier"); !errors.Is(err, ErrFull) {
		t.Errorf("tournoi plein : %v", err)
	}
}
