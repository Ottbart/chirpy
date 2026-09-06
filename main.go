package main

import (
	"io"
	"log"
	"net/http"
)

type apiHandler struct{}

//func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	//add handler for the root path
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))

	mux.HandleFunc("/healthz",
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK\n")
		},
	)
	//start server
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
