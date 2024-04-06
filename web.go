package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
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

var (
	errConversionFailed = errors.New("failed to convert value to integer")
)

// WriteJSON is a helper function which automatically writes JSON formated responses to the client.
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", JSON_ENCODING)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func successfulLoginResponse(w http.ResponseWriter, token string) {
	w.Header().Add("x-jwt-token", token)
	w.WriteHeader(http.StatusOK)
}

// ReadJSON is a helper function which reads the contents of a JSON encoded message into a struct dst.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
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

// toInt is a helper function taking an interface{} value and returning an int representation of it, if possible.
// toInt does not account for all possible conversions, like in spf13/cast lib.
func toInt(val any) (int, error) {
	switch num := val.(type) {
	case int64:
		return int(num), nil
	case int32:
		return int(num), nil
	case float64:
		return int(num), nil
	case float32:
		return int(num), nil
	case string:
		return strconv.Atoi(num)
	default:
		return 0, errConversionFailed
	}
}
