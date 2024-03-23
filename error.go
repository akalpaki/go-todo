package main

import "net/http"

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
