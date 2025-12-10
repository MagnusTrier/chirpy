package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type requestVals struct {
		Body string `json:"body"`
	}

	type returnValsError struct {
		Msg string `json:"error"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		w.WriteHeader(500)

		data, err := json.Marshal(returnValsError{Msg: fmt.Sprintf("Error decoding parameters: %s", err)})
		if err != nil {
			fmt.Print(err)
			return
		}
		w.Write(data)
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(400)
		data, err := json.Marshal(returnValsError{Msg: "chirp is too long"})
		if err != nil {
			w.WriteHeader(500)
			fmt.Print(err)
			return
		}
		w.Write(data)
		return
	}

	notAllowed := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax"}
	cleanedChirp := params.Body

	for _, w := range notAllowed {
		cleanedChirp = strings.ReplaceAll(cleanedChirp, w, "****")
	}

	type returnValsSuccess struct {
		CleanedBody string `json:"cleaned_body"`
	}

	resBody := returnValsSuccess{
		CleanedBody: cleanedChirp,
	}

	data, err := json.Marshal(resBody)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
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

	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	s.ListenAndServe()
}
