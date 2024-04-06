package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test")

	type testCase struct {
		name         string
		data         CreateTodo
		token        *string
		expectedCode int
		expectedErr  apiErrorV2
	}

	tc := []testCase{
		{
			name: "happy path",
			data: CreateTodo{
				UserID: 1,
				Name:   "test",
				Items: []Item{
					{
						ItemNo:  1,
						Content: "test",
						Done:    true,
					},
				},
			},
			token:        makeTestToken(t, "happy path", 1),
			expectedCode: http.StatusOK,
		},
		{
			name: "no user specified",
			data: CreateTodo{
				Name: "test",
				Items: []Item{
					{
						ItemNo:  1,
						Content: "test",
						Done:    true,
					},
				},
			},
			token:        makeTestToken(t, "no user specified", 1),
			expectedCode: 400,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
		{
			name:         "no data",
			token:        makeTestToken(t, "no test token", 1),
			expectedCode: 400,
			expectedErr:  apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
		{
			name: "no token",
			data: CreateTodo{
				UserID: 1,
				Name:   "test",
				Items: []Item{
					{
						ItemNo:  1,
						Content: "test",
						Done:    true,
					},
				},
			},
			token:        nil,
			expectedCode: http.StatusUnauthorized,
			expectedErr: apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Detail: "missing or invalid token",
				Status: http.StatusUnauthorized,
			},
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
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/todos")
		resp, err := makeTestRequest(t, tt.name, url, http.MethodPost, tt.token, tt.data)
		body, err := readTestResponse(t, tt.name, tt.expectedCode, resp, err)
		if err != nil {
			t.Fatalf("test case %s fail, error=%s", tt.name, err.Error())
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
			var resTodo Todo
			if err := json.Unmarshal(body, &resTodo); err != nil {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.data, resTodo)
			}
			if resTodo.UserID != tt.data.UserID || resTodo.Name != tt.data.Name {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.data, resTodo)
			}
		}
	}
}

func TestGetTodo(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test")

	type testCase struct {
		name         string
		token        *string
		expectedCode int
		expected     Todo
		expectedErr  apiErrorV2
	}

	tc := []testCase{
		{
			name:         "happy path",
			token:        makeTestToken(t, "happy path", 1),
			expectedCode: http.StatusOK,
			expected: Todo{
				ID:     1,
				Name:   "test",
				UserID: 1,
			},
		},
		{
			name:         "todo doesn't belong to the user",
			token:        makeTestToken(t, "todo doesn't belong to the user", 2),
			expectedCode: http.StatusForbidden,
			expectedErr: apiErrorV2{
				Type:   errTypeForbidden,
				Title:  errTitleForbidden,
				Status: http.StatusForbidden,
				Detail: "you do not have access to this resource",
			},
		},
		{
			name:         "no token",
			expectedCode: http.StatusUnauthorized,
			expectedErr: apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  errTitleUnauthorized,
				Status: http.StatusUnauthorized,
				Detail: "missing or invalid token",
			},
		},
	}

	testApp, tempF := setupApp(t)
	createTestUser(t, testApp.repository.DB)
	srv := httptest.NewServer(testApp.handler)
	defer t.Cleanup(func() {
		srv.Close()
		tempF.Close()
		os.Remove("todo_test.db")
	})

	for _, tt := range tc {
		url := fmt.Sprintf("%s/%s", srv.URL, "v1/todos/1")
		resp, err := makeTestRequest(t, tt.name, url, http.MethodGet, tt.token, nil)
		body, err := readTestResponse(t, tt.name, tt.expectedCode, resp, err)

		if err != nil {
			t.Fatalf("test case %s fail, error=%s", tt.name, err.Error())
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
			var resTodo Todo
			if err := json.Unmarshal(body, &resTodo); err != nil {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expected, resTodo)
			}
			if resTodo.ID != tt.expected.ID || resTodo.Name != tt.expected.Name || resTodo.UserID != tt.expected.UserID || !reflect.DeepEqual(resTodo.Items, tt.expected.Items) {
				t.Fatalf("test case %s failed, expected=%v, result=%v", tt.name, tt.expected, resTodo)
			}
		}
	}
}
