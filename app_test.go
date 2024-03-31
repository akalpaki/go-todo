package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"io"
// 	"log"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func setupTestServer(t *testing.T) *application {
// 	t.Helper()

// 	cfg := LoadTestConfig()
// 	store := NewStorer(cfg.DB)
// 	runMigration(cfg.DB)

// 	return NewApplication(cfg, store)
// }

// func TestHandleCreateUser(t *testing.T) {
// 	type testCase struct {
// 		name         string
// 		data         User
// 		expected     string
// 		expectedCode int
// 	}

// 	cases := []testCase{
// 		{
// 			name:         "happy path",
// 			data:         User{Email: "test@test.com", Password: "test123"},
// 			expected:     `{"user_id": 2, "email": "test2@test.com, "password": "test123"}`,
// 			expectedCode: http.StatusOK,
// 		},
// 		{
// 			name:         "invalid email",
// 			data:         User{Email: "bad_email", Password: "test123"},
// 			expected:     `{"type": "error:bad_request", "status": 400, "title": "invalid request data", "detail": "invalid data"}`,
// 			expectedCode: http.StatusBadRequest,
// 		},
// 	}

// 	for _, c := range cases {
// 		srv := setupTestServer(t)
// 		SeedUsersToTestDB(srv.storer.DB)

// 		testData, err := json.Marshal(c.data)
// 		if err != nil {
// 			log.Println(err)
// 			t.FailNow()
// 		}
// 		rr := httptest.NewRecorder()
// 		req := httptest.NewRequest(http.MethodPost, "localhost:4555/v1/user", bytes.NewReader(testData))
// 		srv.handleCreateUser(rr, req)

// 		res := rr.Result()
// 		if res.StatusCode != c.expectedCode {
// 			t.Fatalf("expected=%d, result=%d", c.expectedCode, res.StatusCode)
// 		}
// 		body := res.Body
// 		defer body.Close()
// 		data, err := io.ReadAll(body)
// 		if err != nil {
// 			log.Println(err)
// 			t.FailNow()
// 		}
// 		// test happy path, verify registered user is what you expect
// 		if c.expectedCode == http.StatusOK {

// 		}
// 	}
// }
