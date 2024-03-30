package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createAccessToken(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := jwt.MapClaims{
		"iss": "todo",
		"sub": userID,
		"aud": "todo",
		"exp": time.Now().Add(30 * time.Minute),
		"iat": time.Now(),
	}
	token.Claims = claims
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET_TOKEN")
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("incorrect signing method: %s", t.Header["alg"])
		}
		return secret, nil
	})
}

func isExpired(expiresAt time.Time) bool {
	now := time.Now()
	return now.Before(expiresAt)
}
