package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	type testCase struct {
		name           string
		data           CreateTodo
		token          string
		expectedStatus int
		expected       Todo
		expectedErr    apiErrorV2
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
			token:          makeTestToken(t, "happy path"),
			expectedStatus: http.StatusOK,
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
			token:          makeTestToken(t, "no user specified"),
			expectedStatus: 400,
			expectedErr:    apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
		},
		{
			name:           "no data",
			token:          makeTestToken(t, "no test token"),
			expectedStatus: 400,
			expectedErr:    apiErrorV2{Type: errTypeBadRequest, Status: http.StatusBadRequest, Title: errTitleBadRequest, Detail: "invalid data"},
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
			expectedStatus: http.StatusUnauthorized,
			expectedErr: apiErrorV2{
				Type:   errTypeUnauthorized,
				Title:  "Missing or invalid credentials",
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
		resp, err := makeTestRequest(t, tt.name, url, http.MethodPost, tt.data)
		body, err := readTestResponse(t, tt.name, tt.expectedStatus, resp, err)
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
		}
	}
}