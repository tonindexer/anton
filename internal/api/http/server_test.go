package http

import (
	"context"
	"testing"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/query"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _testService *query.Service

var ctx = context.Background()

func testService(t *testing.T) *query.Service {
	if _testService != nil {
		return _testService
	}

	bd, err := repository.ConnectDB(ctx,
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://user:pass@localhost:5432/default?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	s, err := query.NewService(context.Background(), &app.QueryConfig{DB: bd})
	if err != nil {
		t.Fatal(err)
	}

	_testService = s
	return _testService
}

func TestServer_Start(t *testing.T) {
	c := NewController(testService(t))

	s := NewServer(":8080")

	s.RegisterRoutes(c)

	t.Fatal(s.Run())
}
