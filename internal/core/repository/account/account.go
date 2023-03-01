package account

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/addr"
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

func createIndexes(ctx context.Context, pgDB *bun.DB) error {
	// account data

	_, err := pgDB.NewCreateIndex().
		Model(&core.AccountData{}).
		Using("HASH").
		Column("owner_address").
		Where("length(owner_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountData{}).
		Using("HASH").
		Column("collection_address").
		Where("length(collection_address) > 0").
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
		Column("address", "latest").
		Where("latest IS TRUE").
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
		return errors.Wrap(err, "latest account state pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountData{}).
		Using("GIN").
		Column("types").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state contract types pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.AccountState{}).
		Using("BTREE").
		Column("last_tx_lt").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state last_tx_lt pg create index")
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

	return createIndexes(ctx, pgDB)
}

func accountAddresses(accounts []*core.AccountState) (ret []*addr.Address) {
	m := make(map[addr.Address]struct{})
	for _, a := range accounts {
		m[a.Address] = struct{}{}
	}
	for a := range m {
		r := new(addr.Address)
		*r = a
		ret = append(ret, r)
	}
	return
}

func (r *Repository) AddAccountStates(ctx context.Context, tx bun.Tx, accounts []*core.AccountState) error {
	if len(accounts) == 0 {
		return nil
	}

	_, err := r.ch.NewInsert().Model(&accounts).Exec(ctx)
	if err != nil {
		return err
	}

	_, err = tx.NewUpdate().Model(&accounts).
		Where("address in (?)", bun.In(accountAddresses(accounts))).
		Where("latest = ?", true).
		Set("latest = ?", false).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot drop latest state")
	}

	_, err = tx.NewInsert().Model(&accounts).ExcludeColumn("latest").Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot insert new states")
	}

	_, err = tx.NewUpdate().
		With("late",
			tx.NewSelect().
				Model(&accounts).
				Column("address").
				ColumnExpr("max(last_tx_lt) AS max_tx_lt").
				Where("address in (?)", bun.In(accountAddresses(accounts))).
				Group("address"),
		).
		Model((*core.AccountState)(nil)).
		Table("late").
		Where("account_state.address = late.address").
		Where("account_state.last_tx_lt = late.max_tx_lt").
		Set("latest = ?", true).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot set latest state")
	}

	return nil
}

func (r *Repository) AddAccountData(ctx context.Context, tx bun.Tx, data []*core.AccountData) error {
	if len(data) == 0 {
		return nil
	}
	_, err := r.ch.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = tx.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetAccountStates(ctx context.Context, f *core.AccountStateFilter) (ret []*core.AccountState, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.WithData {
		q = q.Relation("StateData")
	}

	q = q.ExcludeColumn("code", "data") // TODO: optional

	if f.LatestState {
		q.Where("account_state.latest = ?", true)
	}
	if len(f.Addresses) > 0 {
		q.Where("account_state.address in (?)", bun.In(f.Addresses))
	}

	if f.WithData {
		if len(f.ContractTypes) > 0 {
			q.Where("state_data.types && ?", pgdialect.Array(f.ContractTypes))
		}
		if f.OwnerAddress != nil {
			q = q.Where("state_data.owner_address = ?", f.OwnerAddress)
		}
		if f.CollectionAddress != nil {
			q = q.Where("state_data.collection_address = ?", f.CollectionAddress)
		}
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where("account_state.last_tx_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where("account_state.last_tx_lt < ?", f.AfterTxLT)
		}
	}

	if f.Order != "" {
		q = q.Order("account_state.last_tx_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)

	return ret, err
}
