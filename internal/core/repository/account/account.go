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
		Column("minter_address").
		Where("length(minter_address) > 0").
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

	// latest account state

	_, err = pgDB.NewCreateIndex().
		Model(&core.LatestAccountState{}).
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

	_, err = pgDB.NewCreateTable().
		Model(&core.LatestAccountState{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "account state pg create table")
	}

	return createIndexes(ctx, pgDB)
}

func (r *Repository) AddAccountStates(ctx context.Context, tx bun.Tx, accounts []*core.AccountState) error {
	if len(accounts) == 0 {
		return nil
	}

	_, err := tx.NewInsert().Model(&accounts).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot insert new states")
	}

	for _, a := range accounts {
		_, err = tx.NewInsert().
			Model(&core.LatestAccountState{
				Address:  a.Address,
				LastTxLT: a.LastTxLT,
			}).
			On("CONFLICT (address) DO UPDATE").
			Where("latest_account_state.last_tx_lt < ?", a.LastTxLT).
			Set("last_tx_lt = EXCLUDED.last_tx_lt").
			Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "cannot set latest state for %s", &a.Address)
		}
	}

	_, err = r.ch.NewInsert().Model(&accounts).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddAccountData(ctx context.Context, tx bun.Tx, data []*core.AccountData) error {
	if len(data) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&data).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func joinLatestAccountData(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.ColumnExpr(
		"\"account_state__state_data\".\"address\" AS \"account_state__state_data__address\", " +
			"\"account_state__state_data\".\"last_tx_lt\" AS \"account_state__state_data__last_tx_lt\", " +
			"\"account_state__state_data\".\"last_tx_hash\" AS \"account_state__state_data__last_tx_hash\", " +
			"\"account_state__state_data\".\"balance\" AS \"account_state__state_data__balance\", " +
			"\"account_state__state_data\".\"types\" AS \"account_state__state_data__types\", " +
			"\"account_state__state_data\".\"owner_address\" AS \"account_state__state_data__owner_address\", " +
			"\"account_state__state_data\".\"minter_address\" AS \"account_state__state_data__minter_address\", " +
			"\"account_state__state_data\".\"next_item_index\" AS \"account_state__state_data__next_item_index\", " +
			"\"account_state__state_data\".\"royalty_address\" AS \"account_state__state_data__royalty_address\", " +
			"\"account_state__state_data\".\"royalty_factor\" AS \"account_state__state_data__royalty_factor\", " +
			"\"account_state__state_data\".\"royalty_base\" AS \"account_state__state_data__royalty_base\", " +
			"\"account_state__state_data\".\"content_uri\" AS \"account_state__state_data__content_uri\", " +
			"\"account_state__state_data\".\"content_name\" AS \"account_state__state_data__content_name\", " +
			"\"account_state__state_data\".\"content_description\" AS \"account_state__state_data__content_description\", " +
			"\"account_state__state_data\".\"content_image\" AS \"account_state__state_data__content_image\", " +
			"\"account_state__state_data\".\"content_image_data\" AS \"account_state__state_data__content_image_data\", " +
			"\"account_state__state_data\".\"initialized\" AS \"account_state__state_data__initialized\", " +
			"\"account_state__state_data\".\"item_index\" AS \"account_state__state_data__item_index\", " +
			"\"account_state__state_data\".\"editor_address\" AS \"account_state__state_data__editor_address\", " +
			"\"account_state__state_data\".\"total_supply\" AS \"account_state__state_data__total_supply\", " +
			"\"account_state__state_data\".\"mintable\" AS \"account_state__state_data__mintable\", " +
			"\"account_state__state_data\".\"admin_address\" AS \"account_state__state_data__admin_address\", " +
			"\"account_state__state_data\".\"jetton_balance\" AS \"account_state__state_data__jetton_balance\", " +
			"\"account_state__state_data\".\"errors\" AS \"account_state__state_data__errors\"",
	)

	q = q.Join("LEFT JOIN account_data AS account_state__state_data ON " +
		"latest_account_state.address = account_state__state_data.address AND " +
		"latest_account_state.last_tx_lt = account_state__state_data.last_tx_lt")

	return q
}

func addAccountDataFilter(q *bun.SelectQuery, f *core.AccountStateFilter) *bun.SelectQuery {
	if !f.WithData {
		return q
	}

	prefix := ""
	if f.LatestState {
		prefix = "account_state__"
	}

	if len(f.ContractTypes) > 0 {
		q = q.Where(prefix+"state_data.types && ?", pgdialect.Array(f.ContractTypes))
	}
	if f.OwnerAddress != nil {
		q = q.Where(prefix+"state_data.owner_address = ?", f.OwnerAddress).
			Where("length(" + prefix + "state_data.owner_address) > 0")
	}
	if f.MinterAddress != nil {
		q = q.Where(prefix+"state_data.minter_address = ?", f.MinterAddress).
			Where("length(" + prefix + "state_data.minter_address) > 0")
	}

	return q
}

func (r *Repository) filterAccountStates(ctx context.Context, f *core.AccountStateFilter) (ret []*core.AccountState, err error) {
	var (
		q           *bun.SelectQuery
		statesTable string
		latest      []*core.LatestAccountState
	)

	// choose table to filter states by
	// and optionally join account data
	if f.LatestState {
		q = r.pg.NewSelect().Model(&latest).Relation("AccountState", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.ExcludeColumn("code", "data") // TODO: optional
		})
		if f.WithData {
			q = joinLatestAccountData(q)
		}
		statesTable = "latest_account_state."
	} else {
		q = r.pg.NewSelect().Model(&ret).
			ExcludeColumn("code", "data") // TODO: optional
		if f.WithData {
			q = q.Relation("StateData")
		}
		statesTable = "account_state."
	}

	if len(f.Addresses) > 0 {
		q = q.Where(statesTable+"address in (?)", bun.In(f.Addresses))
	}

	q = addAccountDataFilter(q, f)

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where(statesTable+"last_tx_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where(statesTable+"last_tx_lt < ?", f.AfterTxLT)
		}
	}
	if f.Order != "" {
		q = q.Order(statesTable + "last_tx_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)

	if f.LatestState {
		for _, a := range latest {
			ret = append(ret, a.AccountState)
		}
	}

	return ret, err
}

func (r *Repository) countAccountStates(ctx context.Context, f *core.AccountStateFilter) (int, error) {
	var data bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if f.WithData {
		if len(f.ContractTypes) > 0 {
			q, data = q.Where("hasAny(types, [?])", ch.In(f.ContractTypes)), true
		}
		if f.OwnerAddress != nil {
			q, data = q.Where("owner_address = ?", f.OwnerAddress), true
		}
		if f.MinterAddress != nil {
			q, data = q.Where("minter_address = ?", f.MinterAddress), true
		}
	}

	if len(f.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(f.Addresses))
	}

	if data {
		q = q.Model((*core.AccountData)(nil))
	} else {
		q = q.Model((*core.AccountState)(nil))
	}

	if f.LatestState {
		q = q.ColumnExpr("argMax(address, last_tx_lt)").
			Group("address")
	} else {
		q = q.ColumnExpr("address")
	}

	return r.ch.NewSelect().TableExpr("(?) as q", q).Count(ctx)
}

func (r *Repository) GetAccountStates(ctx context.Context, f *core.AccountStateFilter) (*core.AccountStateFilterResults, error) {
	var (
		res = new(core.AccountStateFilterResults)
		err error
	)

	res.Rows, err = r.filterAccountStates(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countAccountStates(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}
