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

func defaultKeyFunc(t *jwt.Token) (any, error) {
	if t.Method.Alg() != jwt.SigningMethodHS256 {

	}
}

func Auth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")

		handlerFunc(w, r)
	}
}

// withJWTTodoAuth is authentication middleware for routes handling the todo resource.
func withJWTTodoAuth(handlerFunc http.HandlerFunc, repo *repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		if tokenStr == "" {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		token, err := validateJWT(tokenStr)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		todoID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiErrorV2{
				Type:   errTypeBadRequest,
				Title:  "Invalid Todo ID",
				Status: http.StatusBadRequest,
			})
			return
		}

		todo, err := repo.GetTodoMetadataByID(todoID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiErrorV2{
				Type:   errTypeInternalServerError,
				Title:  "Unable to process your request",
				Status: http.StatusInternalServerError,
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if todo.UserID != int(claims["sub"].(float64)) {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		if claims["iss"] != "todo" {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		if claims["aud"] != "todo" {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}
		exp, err := claims.GetExpirationTime()
		if err != nil || exp == nil {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		if isExpired(exp.Time) {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
				Status: http.StatusUnauthorized,
			})
			return
		}

		handlerFunc(w, r)
	}
}
