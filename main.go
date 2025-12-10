package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

func (cfg *apiConfig) handlerResetFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}

func main() {

	mux := http.NewServeMux()

	cfg := apiConfig{}

	filepath := http.Dir(".")

	customHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	handler := http.StripPrefix("/app/", http.FileServer(filepath))

	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", customHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerFileserverHits)
	mux.HandleFunc("POST /admin/reset", cfg.handlerResetFileserverHits)

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	s.ListenAndServe()
}
