package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akalpaki/todo/internal/app"
)

func main() {
	cfg := loadConfig()

	pool := connectToDatabase(cfg.ConnStr)
	logger := initLogger(cfg.LogLevel, cfg.LoggerOutput)

	app := app.New(cfg, logger, pool)
	httpSrv := http.Server{
		Addr:    cfg.ListenAddr,
		Handler: app,
	}

	log.Println("server running at port ", cfg.ListenAddr)
	log.Fatal(httpSrv.ListenAndServe())
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
	if out != os.Stdout.Name() {
		f, err := os.Open(out)
		if err != nil {
			panic("invalid log output file given!")
		}
		h = slog.NewJSONHandler(f, &slog.HandlerOptions{AddSource: false, Level: level})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: level})
	}

	return slog.New(h)
}
