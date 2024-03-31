package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	Env        string
	ListenAddr string
	DB         *sql.DB
	Logger     *slog.Logger
}

// LoadConfig loads environment variables into config and establishes a database connection
func LoadConfig() *config {
	godotenv.Load(".env")

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		panic("env var ENVIRONMENT must be set")
	}

	listenAddr := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))

	dbConnStr := os.Getenv("DB_CONNECTION_STRING")
	db, err := sql.Open("sqlite3", dbConnStr)
	if err != nil {
		log.Fatalf("failed to open db connection: error=%s", err.Error())
	}

	var logger *slog.Logger
	switch env {
	case "development":
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case "production":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return &config{
		Env:        env,
		ListenAddr: listenAddr,
		DB:         db,
		Logger:     logger,
	}
}

func LoadTestConfig() *config {
	env := "testing"
	listenAddr := ":4555"
	db, err := sql.Open("sqlite3", "file:todo_test.db")
	if err != nil {
		log.Fatalf("failed to open db connection: error=%s", err.Error())
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &config{
		Env:        env,
		ListenAddr: listenAddr,
		DB:         db,
		Logger:     logger,
	}
}
