package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Ottbart/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body        string    `json:"body"`
		UserID      uuid.UUID `json:"user_id"`
		BodyCleaned string    `json:"body_cleaned"`
	}

	// decode json from request
	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("error decoding chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error decoding chirp")
	}

	//check if chirp has max 140 characters
	if len(req.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	//clean text from bad words
	req.BodyCleaned = cleanText(req.Body)

	//write chirp to database
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   req.BodyCleaned,
		UserID: req.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		return
	}

	//response OK
	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
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
