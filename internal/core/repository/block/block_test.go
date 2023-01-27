package block

import (
	"context"
	"reflect"
	"testing"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/db"
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

var _blockRepo *Repository

func blockRepo(t *testing.T) *Repository {
	if _blockRepo != nil {
		return _blockRepo
	}

	_blockRepo = NewRepository(chdb(t))
	return _blockRepo
}

func TestBlockRepository_AddMasterBlock(t *testing.T) {
	mb := &core.BlockInfo{
		Workchain: -1,
		Shard:     1,
		SeqNo:     12,
		RootHash:  core.RandBytes(),
		FileHash:  core.RandBytes(),
	}
	mb2 := &core.BlockInfo{
		Workchain: -1,
		Shard:     1,
		SeqNo:     13,
		RootHash:  core.RandBytes(),
		FileHash:  core.RandBytes(),
	}

	if err := blockRepo(t).AddBlocksInfo(ctx, []*core.BlockInfo{mb}); err != nil {
		t.Fatal(err)
	}
	if err := blockRepo(t).AddBlocksInfo(ctx, []*core.BlockInfo{mb2}); err != nil {
		t.Fatal(err)
	}

	// got, err := blockRepo(t).GetMasterBlockBySeqNo(ctx, mb.SeqNo)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if !reflect.DeepEqual(got, mb) {
	// 	t.Fatalf("different masterchain blocks, expected: %+v, got: %+v", mb, got)
	// }

	got, err := blockRepo(t).GetLastMasterBlockInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, mb2) {
		t.Fatalf("different masterchain blocks, expected: %+v, got: %+v", mb2, got)
	}

	if err := blockRepo(t).AddBlocksInfo(ctx, []*core.BlockInfo{mb}); err != nil {
		t.Fatal(err)
	}
}
