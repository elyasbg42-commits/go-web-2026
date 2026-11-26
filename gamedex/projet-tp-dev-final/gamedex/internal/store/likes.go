// Package store est notre fausse base de données : tout vit en mémoire
// et disparaît quand le serveur s'arrête. Une vraie base (SQLite, PostgreSQL)
// ferait la même chose, mais sur le disque.
package store

import "sync"

// Le serveur traite plusieurs requêtes en même temps.
// Le mutex garantit qu'une seule à la fois touche la map.
var (
	mu    sync.Mutex
	likes = map[int]int{} // numéro du jeu -> nombre de likes
)

// AddLike ajoute un like au jeu et renvoie le nouveau total.
func AddLike(gameID int) int {
	mu.Lock()
	defer mu.Unlock()
	likes[gameID]++
	return likes[gameID]
}

// Likes renvoie le nombre de likes d'un jeu (0 s'il n'en a jamais eu).
func Likes(gameID int) int {
	mu.Lock()
	defer mu.Unlock()
	return likes[gameID]
}

// AllLikes renvoie une copie de tous les compteurs : pour trier ou afficher
// toute une liste sans reprendre le verrou à chaque jeu.
// C'est une copie : la modifier ne change pas la vraie map.
func AllLikes() map[int]int {
	mu.Lock()
	defer mu.Unlock()
	copie := make(map[int]int, len(likes))
	for id, n := range likes {
		copie[id] = n
	}
	return copie
}
