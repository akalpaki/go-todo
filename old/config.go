package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ENV_DEV  = "development"
	ENV_PROD = "production"
)

type config struct {
	Env           string
	ListenAddr    string
	ConnectionStr string // DB connection string
}

func loadCfg() *config {
	err := godotenv.Load(".env")
	if err != nil {
		// defaults
		return &config{
			Env:           ENV_DEV,
			ListenAddr:    ":3000",
			ConnectionStr: "file:todo.db",
		}
	}
	return &config{
		Env:           os.Getenv("ENVIRONMENT"),
		ListenAddr:    fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		ConnectionStr: os.Getenv("DB_CONNECTION_STRING"),
	}
}
