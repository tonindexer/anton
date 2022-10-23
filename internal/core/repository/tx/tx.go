package tx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

type Repository struct {
	db *ch.DB
}

func NewRepository(db *ch.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddTransactions(ctx context.Context, transactions []*core.Transaction) error {
	for _, tx := range transactions {
		_, err := r.db.NewInsert().Model(tx).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetTransactionByHash(ctx context.Context, txHash []byte) (*core.Transaction, error) {
	ret := new(core.Transaction)

	err := r.db.NewSelect().Model(ret).
		Where(fmt.Sprintf("hash = '%s'", txHash)).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *Repository) AddMessages(ctx context.Context, messages []*core.Message) error {
	for _, msg := range messages {
		_, err := r.db.NewInsert().Model(msg).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetMessageByHash(ctx context.Context, msgHash []byte) (*core.Message, error) {
	ret := new(core.Message)

	err := r.db.NewSelect().Model(ret).
		Where("hash = ?", msgHash).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *Repository) GetInMessageByTxHash(ctx context.Context, txHash []byte) (*core.Message, error) {
	ret := new(core.Message)

	tx, err := r.GetTransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	err = r.db.NewSelect().Model(ret).
		Where(fmt.Sprintf("tx_hash = '%s'", txHash)).
		Where(fmt.Sprintf("dst_addr = '%s'", tx.AccountAddr)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *Repository) GetOutMessagesByTxHash(ctx context.Context, txHash []byte) ([]*core.Message, error) {
	ret := new([]*core.Message)

	tx, err := r.GetTransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	err = r.db.NewSelect().Model(ret).
		Where(fmt.Sprintf("tx_hash = '%s'", txHash)).
		Where(fmt.Sprintf("src_addr = '%s'", tx.AccountAddr)).
		Order("created_lt ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return *ret, nil
}

func (r *Repository) AddMessagePayloads(ctx context.Context, payloads []*core.MessagePayload) error {
	for _, msg := range payloads {
		_, err := r.db.NewInsert().Model(msg).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
