package testutils

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func initDatabase(connStr string) *pgxpool.Pool {
	pool := connectToDB(connStr)
	if err := setupTables(pool); err != nil {
		panic("failed to set up database")
	}
	return pool
}

func connectToDB(connStr string) *pgxpool.Pool {
	// TODO: fix error coonecting to db container.
	connCtx := context.TODO()
	pool, err := pgxpool.New(connCtx, connStr)
	if err != nil {
		log.Fatalf("test init: opening db: %s", err.Error())
	}
	if err := pool.Ping(connCtx); err != nil {
		log.Fatalf("test init: ping db: %s", err.Error())
	}

	if err := setupTables(pool); err != nil {
		log.Fatalf("test init: running db setup script: %s", err.Error())
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

	seedData(conn)
	return nil
}

func seedData(conn *pgxpool.Pool) {
	pass1, err := bcrypt.GenerateFromPassword([]byte("test1"), 14)
	if err != nil {
		panic("failed to create test user password")
	}
	pass2, err := bcrypt.GenerateFromPassword([]byte("test2"), 14)
	if err != nil {
		panic("failed to create test user password")
	}
	user1 := fmt.Sprintf(`INSERT INTO users (id, email, password) VALUES ('test1','test1@test.com', %s)`, pass1)
	user2 := fmt.Sprintf(`INSERT INTO users (id, email, password) VALUES ('test2','test2@test.com', %s)`, pass2)

	if _, err := conn.Exec(context.TODO(), user1); err != nil {
		log.Fatalf("failed to seed test user, error=%s", err.Error())
	}
	if _, err := conn.Exec(context.TODO(), user2); err != nil {
		log.Fatalf("failed to seed test user, error=%s", err.Error())
	}

	todo1 := `INSERT INTO todos (id, name, author_id) VALUES ('todo1', 'test1', 'test1)`
	todo2 := `INSERT INTO todos (id, name, author_id) VALUES ('todo2', 'test2', 'test2)`

	if _, err := conn.Exec(context.TODO(), todo1); err != nil {
		log.Fatalf("failed to seed test todo, error=%s", err.Error())
	}
	if _, err := conn.Exec(context.TODO(), todo2); err != nil {
		log.Fatalf("failed to seed test todo, error=%s", err.Error())
	}

	task := `INSERT INTO tasks (id, todo_id, task_order, content, done) VALUES ('task1', 'todo1', 0, TRUE)`

	if _, err := conn.Exec(context.TODO(), task); err != nil {
		log.Fatalf("failed to seed test task, error=%s", err.Error())
	}
}
