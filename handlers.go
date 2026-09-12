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

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		log.Fatalln("can't reset outside PLATFORM=dev")
		respondWithError(w, http.StatusForbidden, "can't reset outside PLATFORM=dev")
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteUser(r.Context())
	if err != nil {
		log.Fatalf("error deleting users from db: %v", err)
		respondWithError(w, 500, "error deleting users from database")
	}
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

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	var req request
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("error decoding email: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error decoding email")
		return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error creating user")
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	if err := respondWithJSON(w, http.StatusCreated, user); err != nil {
		log.Printf("error sending response: %v", err)
	}
}
