package main

import (
	"encoding/json"
	"fmt"
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
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/user")
		resp, err := makeTestRequest(t, tt.name, url, http.MethodPost, tt.data)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		body, err := readTestResponse(t, tt.name, tt.expectedCode, resp, err)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		switch {
		case tt.expectedErr != apiErrorV2{}:
			var resErr apiErrorV2
			if err := json.Unmarshal(body, &resErr); err != nil {
				t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
			}
			if resErr != tt.expectedErr {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expected, resErr)
			}
		default:
			var resUser User
			if err := json.Unmarshal(body, &resUser); err != nil {
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
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/user/login")
		resp, err := makeTestRequest(t, tt.name, url, http.MethodPost, tt.data)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}

		body, err := readTestResponse(t, tt.name, tt.expectedCode, resp, err)
		if err != nil {
			t.Fatalf("test case %s failed, error=%s", tt.name, err.Error())
		}
		switch {
		case tt.expectedErr != apiErrorV2{}:
			var resErr apiErrorV2
			if err := json.Unmarshal(body, &resErr); err != nil {
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
