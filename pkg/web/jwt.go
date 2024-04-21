package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserID contextKey = "userID"

func CreateAccessToken(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "todo",
		"sub": userID,
		"exp": &jwt.NumericDate{Time: time.Now().Add(30 * time.Minute)},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// auth is the middleware that validates jwt tokens.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		token, err := jwt.Parse(tokenStr, defaultKeyFunc)
		if err != nil || token == nil || !token.Valid {
			WriteJSON(
				w,
				r,
				http.StatusUnauthorized,
				ApiError{
					Status:     http.StatusUnauthorized,
					Title:      UnauthorizedTitle,
					Detail:     "missing or invalid token",
					underlying: err,
				},
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			WriteJSON(
				w,
				r,
				http.StatusUnauthorized,
				ApiError{
					Status:     http.StatusUnauthorized,
					Title:      UnauthorizedTitle,
					Detail:     "missing or invalid token",
					underlying: err,
				},
			)
			return
		}

		if !claimsAreValid(claims) {
			WriteJSON(
				w,
				r,
				http.StatusUnauthorized,
				ApiError{
					Status:     http.StatusUnauthorized,
					Title:      UnauthorizedTitle,
					Detail:     "missing or invalid token",
					underlying: err,
				},
			)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			WriteJSON(
				w,
				r,
				http.StatusUnauthorized,
				ApiError{
					Status:     http.StatusUnauthorized,
					Title:      UnauthorizedTitle,
					Detail:     "missing or invalid token",
					underlying: err,
				},
			)
			return
		}
		ctx := context.WithValue(r.Context(), UserID, userID)

		next(w, r.WithContext(ctx))
	}
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
