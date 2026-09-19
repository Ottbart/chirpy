package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ottbart/chirpy/internal/auth"
	"github.com/Ottbart/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
	Token          string    `json:"token"`
	RefreshToken   string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var req request
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding user credentials")
		return
	}
	hashedPw, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error hashing password")
		return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPw,
	})
	if err != nil {
		log.Printf("error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error creating user")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var req request
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding user credentials")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if ok, err := auth.CheckPasswordHash(req.Password, user.HashedPassword); !ok || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	const ExpiresIn = time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.Secret, ExpiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating JWT")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.db.AddRefreshToken(r.Context(), database.AddRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(ExpiresIn),
	})
	if err != nil {
		log.Printf("error creating refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	oldToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}
	userData, err := cfg.db.GetUserFromRefreshToken(r.Context(), oldToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if userData.ExpiresAt.Before(time.Now()) || userData.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "token expired or revoked")
		return
	}

	const ExpiresIn = time.Hour
	token, err := auth.MakeJWT(userData.UserID, cfg.Secret, ExpiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: token})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}
	if err := cfg.db.RevokeToken(r.Context(), token); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error revoking token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateCredentials(w http.ResponseWriter, r *http.Request) {

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

	//get user data from body
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	var req request
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding user credentials")
		return
	}

	//hash new password
	hashedPW, err := auth.HashPassword(req.Password)

	//store new credentials in database
	user, err := cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		ID:             userId,
		Email:          req.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating user credentials")
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
		CreatedAt: user.CreatedAt,
	})
}
