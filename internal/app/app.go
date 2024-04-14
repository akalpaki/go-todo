package app

import (
	"log/slog"
	"net/http"

	"github.com/akalpaki/todo/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akalpaki/todo/internal/todo"
	"github.com/akalpaki/todo/internal/user"
)

func New(
	cfg *config.Config,
	logger *slog.Logger,
	dbPool *pgxpool.Pool,
) http.Handler {
	server := http.NewServeMux()

	userRepo := user.NewRepository(dbPool)
	todoRepo := todo.NewRepository(dbPool)

	server.Handle("/v1/user/", http.StripPrefix("/v1/user", user.Routes(logger, userRepo)))
	server.Handle("/v1/todo/", http.StripPrefix("/v1/todo", todo.Routes(logger, todoRepo)))

	return server
}
