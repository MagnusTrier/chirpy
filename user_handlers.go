package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type requestVals struct {
		Email string `json:"email"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	params := requestVals{}
	if err := decoder.Decode(&params); err != nil {
		returnError(w, err, 500)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	w.WriteHeader(201)
	w.Write(data)
}
