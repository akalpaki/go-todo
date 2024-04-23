package testing

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func CleanupDB(pool *pgxpool.Pool) {
	q := `
	DROP TABLE IF EXISTS users CASCADE;
	DROP TABLE IF EXISTS todos CASCADE;
	DROP TABLE IF EXISTS tasks;
	`
	if _, err := pool.Exec(context.TODO(), q); err != nil {
		panic(err)
	}
}

func initDatabase(connStr string) *pgxpool.Pool {
	pool := connectToDB(connStr)
	CleanupDB(pool)
	if err := setupTables(pool); err != nil {
		log.Fatalf("test init: running db setup script: %s", err.Error())
		CleanupDB(pool)
	}
	return pool
}

func connectToDB(connStr string) *pgxpool.Pool {
	connCtx := context.Background()
	pool, err := pgxpool.New(connCtx, connStr)
	if err != nil {
		log.Fatalf("test init: opening db: %s", err.Error())
	}

	if err := pool.Ping(connCtx); err != nil {
		log.Fatalf("test init: ping db: %s", err.Error())
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
		CONSTRAINT fk_author_id
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
		CONSTRAINT fk_todo_id
			FOREIGN KEY(todo_id)
				REFERENCES todos(id)
				ON DELETE CASCADE
	);
	`
	_, err := conn.Exec(context.TODO(), q)
	if err != nil {
		return err
	}

	if err := seedData(conn); err != nil {
		return err
	}
	return nil
}

func seedData(conn *pgxpool.Pool) error {
	pass1, err := bcrypt.GenerateFromPassword([]byte("test1"), 14)
	if err != nil {
		return err
	}
	pass2, err := bcrypt.GenerateFromPassword([]byte("test2"), 14)
	if err != nil {
		return err
	}
	user1 := fmt.Sprintf(`INSERT INTO users (id, email, password) VALUES ('test1','test1@test.com', '%s')`, string(pass1))
	user2 := fmt.Sprintf(`INSERT INTO users (id, email, password) VALUES ('test2','test2@test.com', '%s')`, string(pass2))

	if _, err := conn.Exec(context.TODO(), user1); err != nil {
		return err
	}
	if _, err := conn.Exec(context.TODO(), user2); err != nil {
		return err
	}

	todo1 := `INSERT INTO todos (id, author_id, name) VALUES ('todo1', 'test1', 'test1')`
	todo2 := `INSERT INTO todos (id, author_id, name) VALUES ('todo2', 'test2', 'test2')`

	if _, err := conn.Exec(context.TODO(), todo1); err != nil {
		return err
	}
	if _, err := conn.Exec(context.TODO(), todo2); err != nil {
		return err
	}

	task := `INSERT INTO tasks (id, todo_id, task_order, content, done) VALUES ('task1', 'todo1', 0, 'test', TRUE)`

	if _, err := conn.Exec(context.TODO(), task); err != nil {
		return err
	}

	return nil
}
