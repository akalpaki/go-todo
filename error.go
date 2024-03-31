package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	errTypeBadRequest          = "error:bad_request"
	errTypeInternalServerError = "error:internal_server_error"
	errTypeNotFound            = "error:not_found"
	errTypeForbidden           = "error:forbidden"
	errTypeUnauthorized        = "error:unauthorized"
)

const (
	errTitleBadRequest          = "invalid request data"
	errTitleInternalServerError = "internal server error"
	errTitleNotFound            = "not found"
	errTitleForbidden           = "forbidden"
	errTitleUnauthorized        = "unauthorized"
)

// apiErrorV2 is a model for an error which can be returned from the REST API that follows the RFC-9457 semantics.
type apiErrorV2 struct {
	// Type is a URI that uniquely identifies the problem type.
	// If the URI is a locator, following the link should provide documentation about how to resolve the error.
	// However, the URI does not have to be resolvable.
	Type string `json:"type"`
	// Status is a number indicating the HTTP status code of the response. This is an optional field.
	Status int `json:"status,omitempty"`
	// Title is a short, human readable description of the problem type. It should not change from occurence to occurence.
	Title string `json:"title,omitempty"`
	// Detail is a human readable explanation specific to an occurence of a problem.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI that identifies the specific occurence of the problem.
	// If the URI is dereferencable, then you should be able to grab problem specific details from it.
	Instance string `json:"instance,omitempty"`
	// underlying error value is used internally for logging/debugging purposes.
	underlying error `json:"-"`
}

func (e *apiErrorV2) Error() string {
	if e.underlying == nil {
		return e.Title
	}
	return e.Title + " : " + e.underlying.Error()
}

func (e *apiErrorV2) responseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while parsing response body: %s", err.Error())
	}
	return body, nil
}

func (e *apiErrorV2) responseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/problem+json",
	}
}

func badRequestResponseV2(message string, err error) *apiErrorV2 {
	return &apiErrorV2{
		Type:       errTypeBadRequest,
		Status:     http.StatusBadRequest,
		Title:      errTitleBadRequest,
		Detail:     message,
		underlying: err,
	}
}

func internalErrorResponseV2(message string, err error) *apiErrorV2 {
	return &apiErrorV2{
		Type:       errTypeInternalServerError,
		Status:     http.StatusInternalServerError,
		Title:      errTitleInternalServerError,
		Detail:     message,
		underlying: err,
	}
}

func notFoundResponseV2() *apiErrorV2 {
	return &apiErrorV2{
		Type:   errTypeNotFound,
		Status: http.StatusNotFound,
		Title:  errTitleNotFound,
	}
}

func forbiddenResponseV2() *apiErrorV2 {
	return &apiErrorV2{
		Type:   errTypeForbidden,
		Status: http.StatusForbidden,
		Title:  errTitleForbidden,
	}
}

func unauthorizedResponseV2() *apiErrorV2 {
	return &apiErrorV2{
		Type:   errTypeUnauthorized,
		Status: http.StatusUnauthorized,
		Title:  errTitleUnauthorized,
	}
}

// apiError allows for bubbling up of errors to the http layer, where we can separate the client error
// from the internal error
type apiError struct {
	Status int
	Msg    string
	err    error
}

func (e apiError) Error() string {
	return e.Msg
}

func badRequestResponse(publicMessage string, privateError error) *apiError {
	return &apiError{
		Status: http.StatusBadRequest,
		Msg:    publicMessage,
		err:    privateError,
	}
}

func internalErrorResponse(publicMessage string, privateError error) *apiError {
	return &apiError{
		Status: http.StatusInternalServerError,
		Msg:    publicMessage,
		err:    privateError,
	}
}
