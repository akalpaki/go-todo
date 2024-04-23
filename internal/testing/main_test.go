package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/akalpaki/todo/internal/todo"
	"github.com/akalpaki/todo/internal/user"
	"github.com/akalpaki/todo/pkg/web"
)

// used to test handling of bcrypt's limitation of password length
const reallyLongPassword = "abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc"

var (
	userRepo *user.Repository
	todoRepo *todo.Repository
	dbPool   *pgxpool.Pool
	logger   *slog.Logger
)

func TestMain(m *testing.M) {
	logger, dbPool = Setup()
	userRepo = user.NewRepository(dbPool)
	todoRepo = todo.NewRepository(dbPool)
	m.Run()
	CleanupDB(dbPool)
	dbPool.Close()
}

func TestRegister(t *testing.T) {
	tc := []struct {
		name               string
		data               user.UserRequest
		expectedResult     user.User
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successfuly create a user",
			data: user.UserRequest{
				Email:    "test3@test.com",
				Password: "test123",
			},
			expectedResult: user.User{
				Email: "test3@test.com",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "no data provided",
			data:               user.UserRequest{},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "no email provided",
			data: user.UserRequest{
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "invalid email provided",
			data: user.UserRequest{
				Email:    "whatevenisthis",
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "no password provided",
			data: user.UserRequest{
				Email: "whoops@ididit.again",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data or malformed json",
		},
		{
			name: "password too long",
			data: user.UserRequest{
				Email:    "test3@test.com",
				Password: reallyLongPassword,
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "httperror:internalerror: failed to create user",
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := TestRequest(t, tt.name, "/", http.MethodPost, "", nil, tt.data)

		user.HandleRegister(logger, userRepo).ServeHTTP(rc, req)

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
			var u user.User
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
		data           user.UserRequest
		expectedResult struct {
			iss string
			sub string
		}
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "user successfully logs in",
			data: user.UserRequest{
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
			data:               user.UserRequest{},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
		{
			name: "wrong password",
			data: user.UserRequest{
				Email:    "test1@test.com",
				Password: "wrong_pass",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
		{
			name: "unregistered user",
			data: user.UserRequest{
				Email:    "idont@exist.com",
				Password: "test1",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "httperror:badrequest: invalid data",
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := TestRequest(t, tt.name, "/login", http.MethodPost, "", nil, tt.data)

		user.HandleLogin(logger, userRepo).ServeHTTP(rc, req)

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

func TestCreate(t *testing.T) {
	tc := []struct {
		name               string
		data               todo.TodoRequest
		userID             string
		expectedResult     todo.Todo
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successfully create a todo",
			data: todo.TodoRequest{
				AuthorID: "test1",
				Name:     "test3",
				Tasks: []todo.Task{
					{
						Content: "test",
						Done:    true,
						Order:   0,
					},
				},
			},
			userID:             "test1",
			expectedStatusCode: http.StatusCreated,
			expectedResult: todo.Todo{
				AuthorID: "test1",
				Name:     "test3",
				Tasks: []todo.Task{
					{
						Content: "test",
						Done:    true,
						Order:   0,
					},
				},
			},
		},
		{
			name: "create a todo for another user",
			data: todo.TodoRequest{
				AuthorID: "test1",
				Name:     "test3",
				Tasks: []todo.Task{
					{
						Content: "test",
						Done:    true,
						Order:   0,
					},
				},
			},
			userID:             "test2",
			expectedStatusCode: http.StatusForbidden,
			expectedError:      "httperror:forbidden: you do not have access to this resource",
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := TestRequest(t, tt.name, "/", http.MethodPost, "", nil, tt.data)
		ctx := context.WithValue(req.Context(), web.UserID, tt.userID)
		todo.HandleCreate(logger, todoRepo).ServeHTTP(rc, req.WithContext(ctx))

		if tt.expectedError != "" {
			var actualError string
			if err := json.Unmarshal(rc.Body.Bytes(), &actualError); err != nil {
				t.Fatalf("test_create: case %s: failed to unmarshall response, error=%s", tt.name, err.Error())
			}
			if tt.expectedError != actualError {
				t.Fatalf("test_create: case %s: expectedError=%v, actualError=%v", tt.name, tt.expectedError, actualError)
			}
		} else {
			res := rc.Result()
			if res.StatusCode != tt.expectedStatusCode {
				t.Fatalf("test_register: case %s: expectedStatusCode=%v, actualStatusCode=%v", tt.name, tt.expectedStatusCode, res.StatusCode)
			}
			var td todo.Todo
			resData, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("test_create: case %s: failed to read response body, error=%s", tt.name, err.Error())
			}
			if err := json.Unmarshal(resData, &td); err != nil {
				t.Fatalf("test_create: case %s: failed to unmarshall response, error=%s", tt.name, err.Error())
			}
			if tt.expectedResult.AuthorID != td.AuthorID && tt.expectedResult.Name != td.Name {
				t.Fatalf("test_register: case %s: expectedResult=%v, actualResult=%v", tt.name, tt.expectedResult, td)
			}
		}
	}
}

func TestGetForUser(t *testing.T) {
	tc := []struct {
		name               string
		userID             string
		expectedResult     []todo.Todo
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:   "user successfully retrieves his todo lists",
			userID: "test1",
			expectedResult: []todo.Todo{
				{
					ID:       "todo1",
					AuthorID: "test1",
					Name:     "test1",
					Tasks: []todo.Task{
						{
							ID:      "task1",
							TodoID:  "todo1",
							Order:   0,
							Content: "test",
							Done:    true,
						},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := TestRequest(t, tt.name, "/", http.MethodGet, "", nil, nil)
		ctx := context.WithValue(req.Context(), web.UserID, tt.userID)
		todo.HandleGetForUser(logger, todoRepo).ServeHTTP(rc, req.WithContext(ctx))
		if rc.Result().StatusCode != tt.expectedStatusCode {
			t.Fatalf("test_getbyuser: case %s: expectedStatusCode=%d, actualStatusCode=%d", tt.name, tt.expectedStatusCode, rc.Result().StatusCode)
		}
	}
}
