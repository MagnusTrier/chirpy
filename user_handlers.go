package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MagnusTrier/chirpy/internal/auth"
	"github.com/MagnusTrier/chirpy/internal/database"
	"github.com/google/uuid"
)

type userPublicInfo struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func returnError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)

	type returnValsError struct {
		Msg string `json:"error"`
	}

	data, err := json.Marshal(returnValsError{Msg: fmt.Sprintf("Error: %v", err)})
	if err != nil {
		fmt.Print(err)
		return
	}
	w.Write(data)
}

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type requestVals struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	createUserArgs := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed,
	}

	user, err := cfg.db.CreateUser(r.Context(), createUserArgs)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	responseData := userPublicInfo{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handlerPostLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type requestVals struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}
	dur := time.Duration(60*60) * time.Second

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		returnError(w, fmt.Errorf("Incorrect email or password"), 401)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		returnError(w, fmt.Errorf("Incorrect email or password"), 401)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, dur)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		returnError(w, err, 500)
		return
	}

	createRefreshTokenArgs := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	if _, err := cfg.db.CreateRefreshToken(r.Context(), createRefreshTokenArgs); err != nil {
		returnError(w, err, 500)
		return
	}

	type responseVals struct {
		userPublicInfo
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	responseData := responseVals{
		userPublicInfo: userPublicInfo{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refresh_token,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerPostRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		returnError(w, err, 401)
	}

	tokenInfo, err := cfg.db.GetToken(r.Context(), token)
	if err != nil {
		returnError(w, err, 401)
		return
	}

	if tokenInfo.Expired {
		if err := cfg.db.RevokeToken(r.Context(), token); err != nil {
			returnError(w, err, 500)
			return
		}
		returnError(w, fmt.Errorf("token exipired"), 401)
		return
	}

	jwtToken, err := auth.MakeJWT(tokenInfo.UserID, cfg.jwtSecret, time.Duration(60*60*time.Second))
	if err != nil {
		returnError(w, err, 500)
		return
	}

	type responseVals struct {
		Token string `json:"token"`
	}

	data, err := json.Marshal(responseVals{Token: jwtToken})
	if err != nil {
		w.WriteHeader(500)
		fmt.Print(err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerPostRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		returnError(w, err, 401)
		return
	}

	if err := cfg.db.RevokeToken(r.Context(), token); err != nil {
		returnError(w, err, 500)
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerPutUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		returnError(w, err, 401)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		returnError(w, err, 401)
		return
	}

	type requestVals struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	updateUserArgs := database.UpdateUserParams{
		ID:             userID,
		HashedPassword: hashed,
		Email:          params.Email,
	}

	user, err := cfg.db.UpdateUser(r.Context(), updateUserArgs)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	responseData := userPublicInfo{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Print(err)
		return
	}

	w.WriteHeader(200)
	w.Write((data))

}
