/*
Package contains test setup functions and utils.
*/
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
)

func setupApp(t *testing.T) (*application, *os.File) {
	t.Helper()
	tempFile, err := os.CreateTemp("", "todo_test.db")
	if err != nil {
		t.Fatalf("test setup failed, error=%s", err.Error())
	}
	conn, err := sql.Open("sqlite3", "file:todo_test.db")
	if err != nil {
		t.Fatalf("test setup failed, error=%s", err.Error())
	}
	if err := conn.Ping(); err != nil {
		t.Fatalf("test setup failed, error=%s", err.Error())
	}
	runMigration(conn)
	testRepo := newRepository(conn)
	testRepo.CreateUser(context.Background(), User{Email: "first@user.com", Password: "pass"})
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	app := newApplication(logger, testRepo)
	app.SetupRoutes()
	return app, tempFile
}

func makeTestRequest(t *testing.T, name string, url string, method string, token *string, data any) (*http.Response, error) {
	t.Helper()
	client := http.Client{}

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	if token != nil {
		req.Header.Add("x-jwt-token", *token)
	}

	resp, err := client.Do(req)
	return resp, err
}

func readTestResponse(t *testing.T, name string, expectedCode int, resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}

	if resp.StatusCode != expectedCode {
		t.Fatalf("test case %s failed, expected code=%d, result=%d", name, expectedCode, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func makeTestToken(t *testing.T, name string, userID int) *string {
	token, err := createAccessToken(userID)
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	return &token
}

func createTestUser(t *testing.T, db *sql.DB) {
	sql := "insert into todo (name, user_id) values (?, ?) "

	_, err := db.Exec(sql, "test", 1)
	if err != nil {
		t.Fatalf("failed to create test user, error=%s", err.Error())
	}
}
