package auth

import (
	"files_server/dir"
	"math/rand"
	"net/http"
)

type AuthStorage struct {
	authStorage map[string]*dir.Dir
}

func New() *AuthStorage {
	return &AuthStorage{authStorage: map[string]*dir.Dir{}}
}

func (authStorage *AuthStorage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := generateToken()
	authStorage.authStorage[token] = dir.New()
	w.Write([]byte(token))
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateToken() string {
	token := make([]rune, 16)
	for i := range token {
		token[i] = letters[rand.Intn(len(letters))]
	}
	return string(token)
}
