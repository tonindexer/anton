package repository

import (
	"context"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core/db"
)

var ctx = context.Background()

var _ch *ch.DB
var _pg *bun.DB

func _initDB(t *testing.T) {
	var err error

	_ch, _pg, err = db.Connect(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
}

func chdb(t *testing.T) *ch.DB {
	if _ch != nil {
		return _ch
	}
	_initDB(t)
	return _ch
}

func pgdb(t *testing.T) *bun.DB {
	if _pg != nil {
		return _pg
	}
	_initDB(t)
	return _pg
}

func TestCreateTables(t *testing.T) {
	err := db.CreateTables(ctx, chdb(t), pgdb(t))
	if err != nil {
		t.Fatal(err)
	}
}
