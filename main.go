package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	apiCfg := apiConfig{}

	//add handler for the root path
	//mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	//mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerReadRequestCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetCounter)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)

	//start server
	log.Println("Serving files from / on port 8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
