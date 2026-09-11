package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Ottbart/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func main() {
	//get data from .env
	godotenv.Load()

	//open db connection
	dbURL := os.Getenv("DB_URL")
	dbConnect, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error open database connection: %v", err)
	}
	//create new database
	dbQueries := database.New(dbConnect)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	//create http.Server
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

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
