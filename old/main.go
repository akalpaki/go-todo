package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DB_SETUP_QUERY = `
create table if not exists todo (id integer not null primary key autoincrement, name text not null, user_id integer not null, foreign key (user_id) references user(id) on delete cascade);
create table if not exists todo_item (itemId integer not null primary key autoincrement, itemNo integer not null, content text not null, done boolean not null, todo_id integer not null, foreign key (todo_id) references todo (id) on delete cascade);
create table if not exists user (id integer not null primary key autoincrement, email text not null, password text not null);
`

func openDBConnection(connStr string) *sql.DB {
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatalf("failed to open database, error=%s", err.Error())
	}

	if err := conn.Ping(); err != nil {
		log.Fatalf("failed to ping database, error=%s", err.Error())
	}

	return conn
}

func setupDB(db *sql.DB) {
	_, err := db.Exec(DB_SETUP_QUERY)
	if err != nil {
		panic("Unable to setup database!")
	}
}

func main() {
	cfg := loadCfg()

	var logger *slog.Logger
	switch cfg.Env {
	case ENV_DEV:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	case ENV_PROD:
		logOutput, err := os.Create("logs.json")
		if err != nil {
			log.Fatalf("unable to setup logger, error=%s", err.Error())
		}
		logger = slog.New(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{AddSource: false, Level: slog.LevelInfo}))
	default:
		log.Fatalln("provided environment is invalid")
	}

	conn := openDBConnection(cfg.ConnectionStr)
	setupDB(conn)

	storer := newRepository(conn)
	app := newApplication(logger, storer)

	srv := http.Server{
		Addr:    cfg.ListenAddr,
		Handler: app.handler,
	}

	log.Println("Server starting at port", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
