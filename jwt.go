package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret string

func init() {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		panic("JWT SECRET KEY NOT SET!")
	}
}

func createAccessToken(userID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := jwt.MapClaims{
		"iss": "todo",
		"sub": userID,
		"aud": "todo",
		"exp": time.Now().Add(30 * time.Minute),
		"iat": time.Now(),
	}
	token.Claims = claims
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func isExpired(expiresAt time.Time) bool {
	now := time.Now()
	return now.Before(expiresAt)
}

func claimsAreValid(claims jwt.MapClaims) bool {
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return false
	}

	if claims["iss"] != "todo" || claims["aud"] != "todo" || isExpired(expiration.Time) {
		return false
	}

	return true
}

func defaultKeyFunc(t *jwt.Token) (any, error) {
	if t.Method.Alg() != jwt.SigningMethodHS256.Name {
		return nil, fmt.Errorf("unexpected singing method: %v", t.Header["alg"])
	}
	return secret, nil
}

// Auth is the middleware that validates jwt tokens.
func Auth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		token, err := jwt.Parse(tokenStr, defaultKeyFunc)
		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			})
		}

		if !claimsAreValid(claims) {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			})
		}

		handlerFunc(w, r)
	}
}
