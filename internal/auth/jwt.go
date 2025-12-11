package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().In(time.UTC)),
			ExpiresAt: jwt.NewNumericDate(time.Now().In(time.UTC).Add(expiresIn)),
			Subject:   userID.String(),
		},
	)

	res, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return res, nil

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	keyF := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	}

	tok, err := jwt.ParseWithClaims(tokenString, &claims, keyF)
	if err != nil {
		return uuid.UUID{}, err
	}
	if !tok.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {

	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("Header did not contain authorization")
	}

	authHeader = strings.TrimPrefix(authHeader, "Bearer ")

	return authHeader, nil
}
