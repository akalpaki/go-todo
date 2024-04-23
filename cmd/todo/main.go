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
	log.Println("Config: ", cfg)
	pool := initDatabase(cfg.ConnStr)
	logger := initLogger(cfg.LogLevel, cfg.LoggerOutput)

	app := app.New(cfg, logger, pool)
	httpSrv := http.Server{
		Addr:    ":8000",
		Handler: app,
	}

	log.Println("server running at port ", cfg.ListenAddr)
	log.Fatal(httpSrv.ListenAndServe())
}

func initDatabase(connStr string) *pgxpool.Pool {
	pool := connectToDB(connStr)
	if err := setupTables(pool); err != nil {
		panic("failed to set up database")
	}
	return pool
}

func connectToDB(connStr string) *pgxpool.Pool {
	connCtx := context.TODO()
	pool, err := pgxpool.New(connCtx, connStr)
	if err != nil {
		log.Fatalf("main: opening db: %s", err.Error())
	}
	if err := pool.Ping(connCtx); err != nil {
		log.Fatalf("main: ping db: %s", err.Error())
	}

	if err := setupTables(pool); err != nil {
		log.Fatalf("main: running db setup script: %s", err.Error())
	}
	return pool
}

func setupTables(conn *pgxpool.Pool) error {
	q := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(21) PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS todos (
		id VARCHAR(21) PRIMARY KEY,
		author_id VARCHAR(21) NOT NULL,
		name TEXT NOT NULL,
		CONSTRAINT author_id
			FOREIGN KEY(author_id)
				REFERENCES users(id)
				ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(21) PRIMARY KEY,
		todo_id VARCHAR(21) NOT NULL,
		content TEXT,
		done BOOLEAN,
		task_order INT,
		CONSTRAINT todo_id
			FOREIGN KEY(todo_id)
				REFERENCES todos(id)
				ON DELETE CASCADE
	);
	`
	_, err := conn.Exec(context.TODO(), q)
	if err != nil {
		return err
	}
	return nil
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
