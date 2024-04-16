package user

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akalpaki/todo/pkg/web"
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
	dbPool.Close()
}

func TestRegister(t *testing.T) {
	tc := []struct {
		name               string
		data               UserRequest
		expectedResult     User
		expectedStatusCode int
		expectedError      *web.ApiError
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
			expectedError: &web.ApiError{
				Status: http.StatusBadRequest,
				Title:  web.BadRequestTitle,
				Detail: "invalid data or malformed json",
			},
		},
		{
			name: "no email provided",
			data: UserRequest{
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError: &web.ApiError{
				Status: http.StatusBadRequest,
				Title:  web.BadRequestTitle,
				Detail: "invalid data or malformed json",
			},
		},
		{
			name: "invalid email provided",
			data: UserRequest{
				Email:    "whatevenisthis",
				Password: "thisisntgonnawork",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError: &web.ApiError{
				Status: http.StatusBadRequest,
				Title:  web.BadRequestTitle,
				Detail: "invalid data or malformed json",
			},
		},
		{
			name: "no password provided",
			data: UserRequest{
				Email: "whoops@ididit.again",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError: &web.ApiError{
				Status: http.StatusBadRequest,
				Title:  web.BadRequestTitle,
				Detail: "invalid data or malformed json",
			},
		},
		{
			name: "password too long",
			data: UserRequest{
				Email:    "test3@test.com",
				Password: reallyLongPassword,
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError: &web.ApiError{
				Status: http.StatusInternalServerError,
				Title:  web.InternalErrorTitle,
				Detail: "failed to create user",
			},
		},
	}

	for _, tt := range tc {
		rc := httptest.NewRecorder()
		req := testutils.TestRequest(t, tt.name, "/", http.MethodPost, nil, nil, tt.data)

		handleRegister(logger, repo).ServeHTTP(rc, req)
		fmt.Println(rc.Result())
	}
}
