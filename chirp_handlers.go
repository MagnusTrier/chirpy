package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MagnusTrier/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type requestVals struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returnValsError struct {
		Msg string `json:"error"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}

	if _, err := cfg.db.GetUser(r.Context(), params.UserID); err != nil {
		returnError(w, err, 500)
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

	chirpArgs := database.CreateChirpParams{
		Body:   cleanedChirp,
		UserID: params.UserID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), chirpArgs)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(500)
		fmt.Print(err)
		return
	}

	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		returnError(w, err, 500)
		return
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(500)
		fmt.Print(err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		returnError(w, err, 404)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		returnError(w, err, 404)
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(500)
		fmt.Print(err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}
