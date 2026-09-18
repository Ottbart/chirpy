package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidToken(t *testing.T) {
	tokenSecret := "tokensecret"
	ExpiresIn := time.Hour
	userID := uuid.New()
	tokenString, _ := MakeJWT(userID, tokenSecret, ExpiresIn)

	uuid, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Errorf("failed ValidateJWT with error: %v", err)
	}
	if uuid != userID {
		t.Errorf("uuid from ValidateJWT %v != userID %v", uuid, userID)
	}
}

func TestInvalidToken(t *testing.T) {
	tokenSecret := "tokensecret"

	_, err := ValidateJWT("invalid.token.string", tokenSecret)
	if err == nil {
		t.Error("test failed, validation with invalid token succeeded")
	}
}

func TestValidBearer(t *testing.T) {
	headers := http.Header{}
	tokenString := "blabla"
	headerToken := "Bearer " + tokenString
	headers.Add("Authorization", headerToken)

	token, _ := GetBearerToken(headers)
	if token != tokenString {
		t.Errorf("token %v from GetBearerToken does not match Auth-Header %v", token, tokenString)
	}
}

func TestInvalidBearer(t *testing.T) {
	headers := http.Header{}
	tokenString := "blabla"
	headerToken := tokenString
	headers.Add("Authorization", headerToken)

	token, err := GetBearerToken(headers)
	if err == nil {
		t.Error("expected GetBearerToken to reject an invalid Authorization header")
	}
	if token != "" {
		t.Errorf("expected no token from invalid Authorization header, got %q", token)
	}
}
