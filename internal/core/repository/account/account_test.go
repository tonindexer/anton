package account_test

import (
	"context"
	"testing"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core/db"
	"github.com/iam047801/tonidx/internal/core/repository/account"
)

var ctx = context.Background()

var _db *ch.DB

func chdb(t *testing.T) *ch.DB {
	if _db != nil {
		return _db
	}

	database, err := db.Connect(context.Background(), "clickhouse://localhost:9000/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	_db = database
	return _db
}

var _addrRepo *account.Repository

func addrRepo(t *testing.T) *account.Repository {
	if _addrRepo != nil {
		return _addrRepo
	}

	_addrRepo = account.NewRepository(chdb(t))
	return _addrRepo
}

func TestRepository_GetContractInterfaces(t *testing.T) {
	ret, err := addrRepo(t).GetContractInterfaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, iface := range ret {
		t.Logf("%+v", iface)
	}
}
