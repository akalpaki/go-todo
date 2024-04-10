package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrInvalidContentType = errors.New("unexpected content type")
	ErrInvalidValue       = errors.New("invalid value")
)

type Validator interface {
	Valid() bool
}

func WriteJSON[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("WriteJSON: %w", err)
	}
	return nil
}

func ReadJSON[T Validator](r *http.Request) (T, error) {
	var v T

	contentType := r.Header.Get("Content-Type")
	if strings.ToLower(contentType) != "application/json" {
		return v, ErrInvalidContentType
	}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("ReadJSON: %w", err)
	}

	if !v.Valid() {
		return v, ErrInvalidValue
	}

	return v, nil
}
