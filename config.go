package main

import (
	"database/sql"
	"log/slog"
)

type config struct {
	ListenAddr string
	DB         *sql.DB
	Logger     *slog.Logger
}
