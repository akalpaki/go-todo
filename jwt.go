package main

import (
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
