package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Ottbart/chirpy/internal/auth"
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

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")

	// If no author_id is provided, retrieve all chirps
	if authorID == "" {
		chirps, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error getting chirps")
			return
		}
		respondWithJSON(w, http.StatusOK, chirps)
	}

	// An author_id was provided: validate and parse it
	authorUuid, err := uuid.Parse(authorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid author ID")
		return
	}
	chirps, err := cfg.db.GetChirpsByAuthor(r.Context(), authorUuid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirps")
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	var chirpID uuid.UUID
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "invalid cirp ID")
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "error getting chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body        string `json:"body"`
		BodyCleaned string `json:"body_cleaned"`
	}

	// check authentication
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}

	// decode json from request
	decoder := json.NewDecoder(r.Body)
	req := request{}
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("error decoding chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error decoding chirp")
		return
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
		UserID: userId,
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	//authenticate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
	}

	userId, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
	}

	//get chirp from request
	var chirpID uuid.UUID
	chirpID, err = uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "invalid cirp ID")
		return
	}

	//get chirp from database and check matching user_id
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp id doesn't exist")
		return
	}

	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "not allowed to delete chirp from other users")
		return
	}

	//delete chirp and respond
	_, err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "can't delete chirp")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
