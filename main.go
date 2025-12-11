package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/MagnusTrier/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Print(err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		db:        dbQueries,
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	filepath := http.Dir(".")

	handler := http.StripPrefix("/app/", http.FileServer(filepath))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerFileserverHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetUsers)

	mux.HandleFunc("GET /api/healthz", handlerHealthz)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handlerPostUsers)

	mux.HandleFunc("POST /api/login", apiCfg.handlerPostLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handlerPostRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerPostRevoke)

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	s.ListenAndServe()
}
