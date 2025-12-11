package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MagnusTrier/chirpy/internal/auth"
	"github.com/MagnusTrier/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type requestVals struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		returnError(w, err, 401)
		return
	}

	if key != cfg.polkaKey {
		returnError(w, fmt.Errorf("invalid api key"), 401)

	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}

	UserID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	if params.Event == "user.upgraded" {
		setUserIsChirpyRedArgs := database.SetUserIsChirpyRedParams{
			ID:          UserID,
			IsChirpyRed: true,
		}
		if err := cfg.db.SetUserIsChirpyRed(r.Context(), setUserIsChirpyRedArgs); err != nil {
			returnError(w, err, 404)
			return
		}
	}

	w.WriteHeader(204)
}
