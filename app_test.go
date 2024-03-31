package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const REALLY_LONG_PASSWORD = "abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc"

func TestCreateUser(t *testing.T) {
	type testCase struct {
		name         string
		data         User
		expected     User
		expectedCode int
		expectedErr  *apiErrorV2
	}

	tc := []testCase{
		{
			name:         "happy path",
			data:         User{Email: "test@test.com", Password: "test123"},
			expected:     User{ID: 2, Email: "test@test.com"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid email",
			data:         User{Email: "nonsense", Password: "test123"},
			expectedErr:  &apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty password",
			data:         User{Email: "test@test.com"},
			expectedErr:  &apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "password too long",
			data:         User{Email: "test@test.com", Password: REALLY_LONG_PASSWORD},
			expectedErr:  &apiErrorV2{Type: errTypeInternalServerError, Status: http.StatusInternalServerError, Title: errTitleInternalServerError, Detail: "failed to create user"},
			expectedCode: http.StatusInternalServerError,
		},
	}

	testApp, tempF := setupApp(t)

	srv := httptest.NewServer(testApp.handler)
	defer srv.Close()

	for _, tt := range tc {
		client := http.Client{}
		reqBody, err := json.Marshal(tt.data)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		req, err := http.NewRequest(http.MethodPost, "localhost:8000/v1/user", bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		if resp.StatusCode != tt.expectedCode {
			t.Fatalf("test case %s failed, expected=%d, result=%d", tt.name, tt.expectedCode, resp.StatusCode)
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		var resUser User
		if err := json.Unmarshal(respBody, &resUser); err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}

	}

	t.Cleanup(func() {
		tempF.Close()
		os.Remove(tempF.Name())
	})
}

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
	testRepo := newRepository(conn)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	app := newApplication(logger, testRepo)
	app.SetupRoutes()
	return app, tempFile
}
