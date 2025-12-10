package main

import (
	"net/http"
)

func main() {

	sm := http.NewServeMux()

	filepath := http.Dir(".")

	customHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.WriteHeader(200)
		w.Write([]byte("OK"))

	}

	// sm.Handle("/app/assets/logo.png", http.StripPrefix("/app/", http.FileServer(filepath)))
	sm.Handle("/app/", http.StripPrefix("/app/", http.FileServer(filepath)))

	sm.HandleFunc("/healthz", customHandler)

	s := http.Server{
		Handler: sm,
		Addr:    ":8080",
	}

	s.ListenAndServe()
}
