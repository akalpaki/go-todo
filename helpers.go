package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const (
	JSON_ENCODING = "application/json"
)

var (
	ErrWrongContentType   = errors.New("incorrect content type")
	ErrUnknownFieldType   = errors.New("unknown field type")
	ErrTooLarge           = errors.New("request body too large")
	ErrTooManyJSONObjects = errors.New("too many json objects")
)

// WriteJSON is a helper function which automatically writes JSON formated responses to the client.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// ReadJSON is a helper function which reads the contents of a JSON encoded message into a struct dst.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	ct := r.Header.Get("Content-Type")
	if ct != JSON_ENCODING {
		return ErrWrongContentType
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		errStr := err.Error()
		switch {
		case strings.HasPrefix(errStr, "json: unknown field "):
			return ErrUnknownFieldType
		case errStr == "http: request body too large":
			return ErrTooLarge
		default:
			return err
		}
	}

	// Decode a second time to ensure all of the request body has been read and doesn't contain another JSON object.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return ErrTooManyJSONObjects
	}

	return nil
}
