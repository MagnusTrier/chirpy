package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/MagnusTrier/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)

	n := cfg.fileserverHits.Load()
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", n)

	w.Write([]byte(body))
}

func (cfg *apiConfig) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	if err := cfg.db.DeleteAllUsers(r.Context()); err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}
