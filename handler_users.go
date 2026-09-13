package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
