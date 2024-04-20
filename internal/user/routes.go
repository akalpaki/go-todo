package user

import (
	"log/slog"
	"net/http"

	"github.com/akalpaki/todo/pkg/web"
)

func Routes(logger *slog.Logger, repository *Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", web.Access(handleRegister(logger, repository), logger))
	mux.HandleFunc("POST /login", web.Access(handleLogin(logger, repository), logger))

	return mux
}

func handleRegister(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data, err := web.ReadJSON[UserRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		user, err := repository.Register(ctx, data)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to create user", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, user); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func handleLogin(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := web.ReadJSON[UserRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data", err)
			return
		}

		registered, err := repository.GetByEmail(ctx, user.Email)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data", err)
			return
		}

		if !passwordMatches(user.Password, registered.Password) {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data", err)
			return
		}

		token, err := web.CreateAccessToken(registered.ID)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data", err)
			return
		}

		w.Header().Add("x-jwt-token", token)
		if err := web.WriteJSON(w, r, http.StatusOK, ""); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data", err)
			return
		}
	}
}
