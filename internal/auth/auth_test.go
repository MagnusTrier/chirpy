package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "this is my secret"
	exp := time.Minute

	token, err := MakeJWT(userID, secret, exp)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("MakeJWT returned empty token")
	}

	jwtID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	if jwtID != userID {
		t.Errorf("expected userID %v, got %v", userID, jwtID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "this is my secret"
	exp := -time.Minute

	token, err := MakeJWT(userID, secret, exp)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}
func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "this is my secret"
	wrongSecret := "WRONG"
	exp := time.Minute

	token, err := MakeJWT(userID, secret, exp)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatalf("expected error for token validation with wrong secret, got nil")
	}
}

func TestGetBearerToken_Valid(t *testing.T) {
	h := http.Header{}
	tok := "reee"
	h.Set("Authorization", "Bearer "+tok)

	token, err := GetBearerToken(h)
	if err != nil {
		t.Fatalf("GetBearerToken returned error: %v", err)
	}

	if token != tok {
		t.Fatalf("expected token %v, got %v", tok, token)
	}
}

func TestGetBearerToken_Missing(t *testing.T) {
	h := http.Header{}

	_, err := GetBearerToken(h)
	if err == nil {
		t.Fatalf("expected missing auth header error, got nil")
	}
}
