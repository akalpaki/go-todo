package user

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akalpaki/todo/internal/testutils"
)

// used to test handling of bcrypt's limitation of password length
const reallyLongPassword = "abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc"

var (
	repo   *Repository
	dbPool *pgxpool.Pool
	logger *slog.Logger
)

func TestMain(m *testing.M) {
	logger, dbPool = testutils.Setup()
	repo = NewRepository(dbPool)
	m.Run()
	testutils.CleanupDB(dbPool)
	dbPool.Close()
}

func TestRegister(t *testing.T) {
	tc := []struct {
		name               string
		data               UserRequest
		expectedResult     User
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successfuly create a user",
			data: UserRequest{
				Email:    "test3@test.com",
				Password: "test123",
			},
			expectedResult: User{
				Email: "test3@test.com",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "no data provided",
			data:               UserRequest{},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "no email provided",
			data: UserRequest{
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "invalid email provided",
			data: UserRequest{
				Email:    "whatevenisthis",
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "no password provided",
			data: UserRequest{
				Email: "whoops@ididit.again",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "password too long",
			data: UserRequest{
				Email:    "test3@test.com",
				Password: reallyLongPassword,
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "httperror:internalerror: failed to create user",
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := testutils.TestRequest(t, tt.name, "/", http.MethodPost, nil, nil, tt.data)

		handleRegister(logger, repo).ServeHTTP(rc, req)

		if tt.expectedStatusCode != rc.Code {
			t.Fatalf("test_register: case %s: expectedStatusCode=%d, actualStatusCode=%d", tt.name, tt.expectedStatusCode, rc.Code)
		}

		if tt.expectedError != "" {
			var actualError string
			if err := json.Unmarshal(rc.Body.Bytes(), &actualError); err != nil {
				t.Fatalf("test_register: case %s: failed to unmarshall response, error=%s", tt.name, err.Error())
			}
			if tt.expectedError != actualError {
				t.Fatalf("test_register: case %s: expectedError=%v, actualError=%v", tt.name, tt.expectedError, actualError)
			}
		} else {
			var u User
			if err := json.Unmarshal(rc.Body.Bytes(), &u); err != nil {
				t.Fatalf("test_register: case %s: failed to unmarshall response, error=%s", tt.name, err.Error())
			}
			if tt.expectedResult.Email != u.Email {
				t.Fatalf("test_register: case %s: expectedResult=%v, actualResult=%v", tt.name, tt.expectedResult, u)
			}
		}
	}
}

func TestLogin(t *testing.T) {
	tc := []struct {
		name           string
		data           UserRequest
		expectedResult struct {
			iss string
			sub string
		}
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "user successfully logs in",
			data: UserRequest{
				Email:    "test1@test.com",
				Password: "test1",
			},
			expectedResult: struct {
				iss string
				sub string
			}{
				iss: "todo",
				sub: "test1",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "no data",
			data:               UserRequest{},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
		{
			name: "wrong password",
			data: UserRequest{
				Email:    "test1@test.com",
				Password: "wrong_pass",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
		{
			name: "unregistered user",
			data: UserRequest{
				Email:    "idont@exist.com",
				Password: "test1",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := testutils.TestRequest(t, tt.name, "/login", http.MethodPost, nil, nil, tt.data)

		handleLogin(logger, repo).ServeHTTP(rc, req)

		if tt.expectedStatusCode != rc.Code {
			t.Fatalf("test_login: case %s: expectedStatusCode=%d, actualStatusCode=%d", tt.name, tt.expectedStatusCode, rc.Code)
		}

		if tt.expectedError != "" {
			var actualError string
			if err := json.Unmarshal(rc.Body.Bytes(), &actualError); err != nil {
				t.Fatalf("test_login: case %s: failed to unmarshall response, error=%s", tt.name, err.Error())
			}
			if tt.expectedError != actualError {
				t.Fatalf("test_login: case %s: expectedError=%v, actualError=%v", tt.name, tt.expectedError, actualError)
			}
		} else {
			tokenStr := rc.Result().Header.Get("x-jwt-token")
			if tokenStr == "" {
				t.Fatalf("test_login: case %s: expectedResult=%s, actualResult=%s", tt.name, tt.expectedResult, tokenStr)
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				secret := os.Getenv("JWT_SECRET_KEY")
				if t.Method.Alg() != jwt.SigningMethodHS256.Name {
					return nil, fmt.Errorf("unexpected singing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			})
			if err != nil {
				t.Fatalf("test_login: case %s: failed to parse token: %s", tt.name, err.Error())
			}

			tokenFields, ok := (token.Claims).(jwt.MapClaims)
			if !ok {
				t.Fatalf("test_login: case %s: token claims are not jwt.MapClaims", tt.name)
			}

			if tokenFields["iss"] != tt.expectedResult.iss && tokenFields["sub"] != tt.expectedResult.sub {
				t.Fatalf("test_login: case %s: expectedResult=%v, actualResult=%v", tt.name, tt.expectedResult, token)
			}
		}
	}
}
