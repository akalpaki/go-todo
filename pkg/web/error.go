package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type title string

const (
	BadRequestTitle       = "httperror:badrequest"
	UnauthorizedTitle     = "httperror:unauthorized"
	ForbiddenTitle        = "httperror:forbidden"
	InternalErrorTitle    = "httperror:internalerror"
	NotFoundTitle         = "httperror:notfound"
	UnspecifiedErrorTitle = "httperror:unspecifiederror"
)

var ErrInvalidUserID = errors.New("invalid user ID")

// ApiError is partial implementation of RFC-9457
type ApiError struct {
	Status     int    `json:"status"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	underlying error  `json:"-"`
}

// Error returns the user-appropriate version of the error
func (e *ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

func ErrorResponse(
	logger *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	status int,
	detail string,
	err error,
) {
	var apiError ApiError
	switch status {
	case http.StatusBadRequest:
		apiError = ApiError{
			Status:     status,
			Title:      BadRequestTitle,
			Detail:     detail,
			underlying: err,
		}
	case http.StatusUnauthorized:
		apiError = ApiError{
			Status:     status,
			Title:      UnauthorizedTitle,
			Detail:     detail,
			underlying: err,
		}
	case http.StatusForbidden:
		apiError = ApiError{
			Status:     status,
			Title:      ForbiddenTitle,
			Detail:     detail,
			underlying: err,
		}
	case http.StatusNotFound:
		apiError = ApiError{
			Status:     status,
			Title:      NotFoundTitle,
			Detail:     detail,
			underlying: err,
		}
	case http.StatusInternalServerError:
		apiError = ApiError{
			Status:     status,
			Title:      InternalErrorTitle,
			Detail:     detail,
			underlying: err,
		}
	default:
		apiError = ApiError{
			Status:     http.StatusInternalServerError,
			Title:      UnspecifiedErrorTitle,
			Detail:     "error details not specified; look into this!",
			underlying: err,
		}
	}
	logger.Error("error", "error_message", apiError.underlying)
	WriteJSON(w, r, apiError.Status, apiError.Error())
}
