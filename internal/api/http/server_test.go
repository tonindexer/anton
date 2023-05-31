package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

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
		"clickhouse://user:pass@localhost:9000/postgres?sslmode=disable",
		"postgres://user:pass@localhost:5432/postgres?sslmode=disable")
	assert.Nil(t, err)

	s, err := query.NewService(context.Background(), &app.QueryConfig{DB: bd})
	assert.Nil(t, err)

	_testService = s
	return _testService
}

func TestServer_Start(t *testing.T) {
	c := NewController(testService(t))

	s := NewServer(":8080")

	s.RegisterRoutes(c)

	assert.Nil(t, s.Run())
}
