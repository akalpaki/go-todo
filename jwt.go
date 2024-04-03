package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createAccessToken(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
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

// withJWTTodoAuth is authentication middleware for routes handling the todo resource.
func withJWTTodoAuth(handlerFunc http.HandlerFunc, repo *repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		if tokenStr == "" {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		token, err := validateJWT(tokenStr)
		if err != nil {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		todoID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		todo, err := repo.GetTodoMetadataByID(todoID)
		if err != nil {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if todo.UserID != int(claims["sub"].(float64)) {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if claims["iss"] != "todo" {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if claims["aud"] != "todo" {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}
		exp, err := claims.GetExpirationTime()
		if err != nil || exp == nil {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if isExpired(exp.Time) {
			writeJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		handlerFunc(w, r)
	}
}
