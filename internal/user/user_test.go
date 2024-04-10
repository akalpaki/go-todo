package user

// TODO: fix this

// import (
// 	"net/http"
// 	"testing"

// 	"github.com/noquark/nanoid"

// 	"github.com/akalpaki/todo/pkg/web"
// )

// const reallyLongPassword = "abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc"

// func TestRegister(t *testing.T) {
// 	expectedID, err := nanoid.New(0)
// 	if err != nil {
// 		t.Fatalf("id generation failed, %s", err.Error())
// 	}
// 	type testCase struct {
// 		name           string
// 		data           UserRequest
// 		expectedResult User
// 		expectedStatus int
// 		expectedErr    web.ApiError
// 	}

// 	tc := []testCase{
// 		{
// 			name:           "happy path",
// 			data:           UserRequest{Email: "test@test.com", Password: "test123"},
// 			expectedResult: User{ID: expectedID, Email: "test@test.com"},
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name:           "invalid email",
// 			data:           UserRequest{Email: "nonsense", Password: "test123"},
// 			expectedErr:    web.ApiError{Status: http.StatusBadRequest, Title: web.BadRequestTitle, Detail: "invalid data or malformed json"},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name:           "empty password",
// 			data:           UserRequest{Email: "test@test.com"},
// 			expectedErr:    web.ApiError{Status: http.StatusBadRequest, Title: web.BadRequestTitle, Detail: "invalid data or malformed json"},
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name:           "password too long",
// 			data:           UserRequest{Email: "test@test.com", Password: reallyLongPassword},
// 			expectedErr:    web.ApiError{Status: http.StatusInternalServerError, Title: web.InternalErrorTitle, Detail: "failed to create user"},
// 			expectedStatus: http.StatusInternalServerError,
// 		},
// 	}
// }
