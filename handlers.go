package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) //optional, because w.Write() would do it if not called before
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerReadRequestCount(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	site := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, count)
	w.Write([]byte(site))
}

func (cfg *apiConfig) handlerResetCounter(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset to 0"))
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body        string `json:"body"`
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}
	// decode json from request
	decoder := json.NewDecoder(r.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("error decoding chirp: %v", err)
		err = respondWithError(w, 500, "error decoding chirp")
		if err != nil {
			log.Printf("error sending error-response: %v", err)
		}
	}

	//check if chirp has max 140 characters
	if len(chirp.Body) > 140 {
		err = respondWithError(w, 400, "Chirp is too long")
		if err != nil {
			log.Printf("error sending error-response: %v", err)
		}
		return
	}
	chirp.CleanedBody = cleanText(chirp.Body)
	chirp.Valid = true

	//response OK
	err = respondWithJSON(w, 200, chirp)
	if err != nil {
		log.Printf("error sending response: %v", err)
	}
}

func cleanText(m string) string {
	badWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	words := strings.Split(m, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
