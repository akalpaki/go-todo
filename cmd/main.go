package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/akalpaki/todo/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := loadConfig()

	pool := connectToDatabase(cfg.ConnStr)

	logger := initLogger(cfg.LogLevel, cfg.LoggerOutput)

	_ = server.New(cfg, logger, pool)
}

func connectToDatabase(connStr string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("main: opening db: %s", err.Error())
	}
	defer pool.Close()

	return pool
}

func initLogger(level slog.Level, out string) *slog.Logger {
	var h slog.Handler
	h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level})
	if out != os.Stdout.Name() {
		f, err := os.Open(out)
		if err != nil {
			panic("invalid log output file given!")
		}
		h = slog.NewJSONHandler(f, &slog.HandlerOptions{AddSource: false, Level: level})
	}
	return slog.New(h)
}
