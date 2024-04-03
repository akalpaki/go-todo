package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
		expectedErr  apiErrorV2
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
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty password",
			data:         User{Email: "test@test.com"},
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "password too long",
			data:         User{Email: "test@test.com", Password: REALLY_LONG_PASSWORD},
			expectedErr:  apiErrorV2{Type: errTypeInternalServerError, Status: http.StatusInternalServerError, Title: errTitleInternalServerError, Detail: "failed to create user"},
			expectedCode: http.StatusInternalServerError,
		},
	}

	testApp, tempF := setupApp(t)

	srv := httptest.NewServer(testApp.handler)
	defer t.Cleanup(func() {
		srv.Close()
		tempF.Close()
		os.Remove("todo_test.db")
	})

	for _, tt := range tc {
		client := http.Client{}
		reqBody, err := json.Marshal(tt.data)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/user")
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		req.Header.Set("Content-Type", "application/json")
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
		switch {
		case tt.expectedErr != apiErrorV2{}:
			var resErr apiErrorV2
			if err := json.Unmarshal(respBody, &resErr); err != nil {
				t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
			}
			if resErr != tt.expectedErr {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expected, resErr)
			}
		default:
			var resUser User
			if err := json.Unmarshal(respBody, &resUser); err != nil {
				t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
			}
			if resUser.Email != tt.expected.Email {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expected.Email, resUser.Email)
			}
			if !checkPasswordHash(tt.data.Password, resUser.Password) {
				t.Fatalf("test case %s failed, passwords do not match", tt.name)
			}
		}

	}
}

func TestLogin(t *testing.T) {
	type testCase struct {
		name         string
		data         User
		expectedCode int
		expectedResp string // sucessful request will send a token
		expectedErr  apiErrorV2
	}
	tc := []testCase{
		{
			name: "happy path",
			data: User{
				Email:    "first@user.com",
				Password: "pass",
			},
			expectedCode: http.StatusOK,
			expectedResp: ``,
		},
		{
			name: "user doesn't exist",
			data: User{
				Email:    "test2@test.com",
				Password: "test123",
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		}, {
			name: "invalid email",
			data: User{
				Email:    "invalid",
				Password: "test123",
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
		{
			name: "no password",
			data: User{
				Email: "test2@test.com",
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
		{
			name: "password too long",
			data: User{
				Email:    "test2@test.com",
				Password: REALLY_LONG_PASSWORD,
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
	}

	testApp, tempF := setupApp(t)

	srv := httptest.NewServer(testApp.handler)
	defer t.Cleanup(func() {
		srv.Close()
		tempF.Close()
		os.Remove("todo_test.db")
	})

	for _, tt := range tc {
		client := http.Client{}
		reqBody, err := json.Marshal(tt.data)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/user/login")
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		req.Header.Set("Content-Type", "application/json")
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
		switch {
		case tt.expectedErr != apiErrorV2{}:
			var resErr apiErrorV2
			if err := json.Unmarshal(respBody, &resErr); err != nil {
				t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
			}
			if resErr != tt.expectedErr {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expectedErr, resErr)
			}
		default:
			token := resp.Header.Get("x-jwt-token")
			if token == "" {
				t.Fatalf("test case %s failed, expected token but received nothing", tt.name)
			}
		}

	}
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
	runMigration(conn)
	testRepo := newRepository(conn)
	testRepo.CreateUser(context.Background(), User{Email: "first@user.com", Password: "pass"})
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	app := newApplication(logger, testRepo)
	app.SetupRoutes()
	return app, tempFile
}
