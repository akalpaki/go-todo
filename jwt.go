package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const USER_ID contextKey = "userID"

func createAccessToken(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := jwt.MapClaims{
		"iss": "todo",
		"sub": userID,
		"exp": &jwt.NumericDate{Time: time.Now().Add(30 * time.Minute)},
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
	return now.After(expiresAt)
}

func claimsAreValid(claims jwt.MapClaims) bool {
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return false
	}

	if claims["iss"] != "todo" || isExpired(expiration.Time) {
		return false
	}

	return true
}

func defaultKeyFunc(t *jwt.Token) (any, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if t.Method.Alg() != jwt.SigningMethodHS256.Name {
		return nil, fmt.Errorf("unexpected singing method: %v", t.Header["alg"])
	}
	return []byte(secret), nil
}

// auth is the middleware that validates jwt tokens.
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		token, err := jwt.Parse(tokenStr, defaultKeyFunc)

		if err != nil || token == nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:       errTypeUnauthorized,
				Title:      errTitleUnauthorized,
				Status:     http.StatusUnauthorized,
				Detail:     "missing or invalid token",
				underlying: err,
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			})
			return
		}

		if !claimsAreValid(claims) {
			writeJSON(w, http.StatusUnauthorized, apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), USER_ID, claims["sub"])

		r = r.WithContext(ctx)

		next(w, r)
	}
}
