package testutils

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/akalpaki/todo/pkg/web"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDBConnString = "host=test-db user=test password=test dbname=test sslmode=disable"

func Setup() (*slog.Logger, *pgxpool.Pool) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	conn := initDatabase(testDBConnString)
	return logger, conn
}

func TestRequest(
	t *testing.T,
	name, url, method string,
	token *string,
	queryParams map[string]string,
	data any,
) *http.Request {
	t.Helper()

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	if queryParams != nil {
		query := req.URL.Query()
		for key, val := range queryParams {
			query.Add(key, val)
		}
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Add("Content-Type", "application/json")
	if token != nil {
		req.Header.Add("x-jwt-token", *token)
	}

	return req
}

func TestToken(t *testing.T, name string, userID string) string {
	token, err := web.CreateAccessToken(userID)
	if err != nil {
		t.Fatalf("test case %s failed, error=%s", name, err.Error())
	}
	return token
}
