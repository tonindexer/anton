package account

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.AccountRepository = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(_ch *ch.DB, _pg *bun.DB) *Repository {
	return &Repository{ch: _ch, pg: _pg}
}

func createIndexes(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	// account data

	_, err := pgDB.NewCreateIndex().
		Model(&core.AccountData{}).
		Unique().
		Column("address", "state_hash").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	// account state

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Using("HASH").
		Column("address").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Unique().
		Column("address", "state_hash").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Unique().
		Column("latest", "address").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Column("latest").
		Where("latest IS TRUE").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state contract types pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Using("GIN").
		Column("types").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state contract types pg create index")
	}

	return nil
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := pgDB.ExecContext(ctx, "CREATE TYPE account_status AS ENUM (?, ?, ?, ?)",
		core.Uninit, core.Active, core.Frozen, core.NonExist)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "account status pg create enum")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.AccountData{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account data ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.AccountData{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account data pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.AccountState{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.AccountState{}).
		IfNotExists().
		// WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state pg create table")
	}

	return createIndexes(ctx, chDB, pgDB)
}

func accountAddresses(accounts []*core.AccountState) (ret []string) {
	for _, a := range accounts {
		ret = append(ret, a.Address)
	}
	return
}

func (r *Repository) AddAccountStates(ctx context.Context, accounts []*core.AccountState) error {
	_, err := r.ch.NewInsert().Model(&accounts).Exec(ctx)
	if err != nil {
		return err
	}

	tx, err := r.pg.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NewUpdate().Model(&accounts).
		Where("address in (?)", bun.In(accountAddresses(accounts))).
		Where("latest = ?", true).
		Set("latest = ?", false).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot drop latest state")
	}

	_, err = tx.NewInsert().Model(&accounts).Exec(ctx)
	if err != nil {
		return err
	}

	_, err = tx.NewUpdate().
		With("late",
			tx.NewSelect().
				Model(&accounts).
				Column("address").
				ColumnExpr("max(last_tx_lt) AS _lt").
				Where("address in (?)", bun.In(accountAddresses(accounts))).
				Group("address"),
		).
		Model((*core.AccountState)(nil)).
		Table("account_states", "late").
		Where("account_states.address = late.address").
		Where("account_states.last_tx_lt = late._lt").
		Set("latest = ?", true).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot set latest state")
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddAccountData(ctx context.Context, data []*core.AccountData) error {
	_, err := r.ch.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.pg.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func selectAccountStatesFilter(q *bun.SelectQuery, filter *core.AccountStateFilter) *bun.SelectQuery {
	if filter.LatestState {
		q.Where("latest = ?", true)
	}
	if filter.Address != "" {
		q.Where("address = ?", filter.Address)
	}
	if len(filter.ContractTypes) > 0 {
		q.Where("contract_types && ?", pgdialect.Array(filter.ContractTypes))
	}
	if filter.WithData {
		if filter.OwnerAddress != "" {
			q = q.Where("state_data.owner_address = ?", filter.OwnerAddress)
		}
		if filter.CollectionAddress != "" {
			q = q.Where("state_data.collection_address = ?", filter.CollectionAddress)
		}
	}
	return q
}

func (r *Repository) GetAccountStates(ctx context.Context, filter *core.AccountStateFilter, offset int, limit int) ([]*core.AccountState, error) {
	var ret []*core.AccountState

	q := r.pg.NewSelect().Model(&ret)
	if filter.WithData {
		q = q.Relation("StateData")
	}

	err := selectAccountStatesFilter(q, filter).
		Order("last_tx_lt DESC").
		Offset(offset).Limit(limit).Scan(ctx)

	return ret, err
}
