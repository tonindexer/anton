package tx

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
)

var _ core.TxRepository = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(_ch *ch.DB, _pg *bun.DB) *Repository {
	return &Repository{ch: _ch, pg: _pg}
}

func createIndexes(ctx context.Context, pgDB *bun.DB) error {
	var err error

	// transactions

	_, err = pgDB.NewCreateIndex().
		Model(&core.Transaction{}).
		Using("HASH").
		Column("address").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction address pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Transaction{}).
		Unique().
		Column("address", "created_lt").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction account lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Transaction{}).
		Column("block_workchain", "block_shard", "block_seq_no").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "tx block id pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Transaction{}).
		Using("BTREE").
		Column("created_lt").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "tx created_lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Transaction{}).
		Using("HASH").
		Column("in_msg_hash").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "tx in_msg hash pg create index")
	}

	// messages

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Column("src_address", "source_tx_lt").
		Where("length(src_address) > 0").
		Where("source_tx_lt > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src_address source_tx_lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("src_address").
		Where("length(src_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Unique().
		Column("src_address", "created_lt").
		Where("length(src_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr lt pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("dst_address").
		Where("length(dst_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message dst addr pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("BTREE").
		Column("created_lt").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message created_lt pg create index")
	}

	// message payloads

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("src_contract").
		Where("length(src_contract) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create src_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("dst_contract").
		Where("length(dst_contract) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create dst_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("operation_name").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create operation name index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.MessagePayload{}).
		Using("HASH").
		Column("minter_address").
		Where("length(minter_address) > 0").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "address state pg create unique index")
	}

	return nil
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := pgDB.ExecContext(ctx, "CREATE TYPE message_type AS ENUM (?, ?, ?)",
		core.ExternalIn, core.ExternalOut, core.Internal)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "account status pg create enum")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.MessagePayload{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.MessagePayload{}).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Message{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Message{}).
		IfNotExists().
		// WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message pg create table")
	}

	_, err = chDB.NewCreateTable().
		IfNotExists().
		Engine("ReplacingMergeTree").
		Model(&core.Transaction{}).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction ch create table")
	}

	_, err = pgDB.NewCreateTable().
		Model(&core.Transaction{}).
		IfNotExists().
		// WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction pg create table")
	}

	if err := createIndexes(ctx, pgDB); err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddTransactions(ctx context.Context, tx bun.Tx, transactions []*core.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&transactions).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&transactions).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddMessages(ctx context.Context, tx bun.Tx, messages []*core.Message) error {
	var unknown []*core.Message

	for _, msg := range messages {
		if msg.Known {
			continue
		}
		unknown = append(unknown, msg)
	}

	if len(unknown) == 0 {
		return nil
	}

	_, err := tx.NewInsert().Model(&unknown).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&unknown).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*core.MessagePayload) error {
	if len(payloads) == 0 {
		return nil
	}
	_, err := tx.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	_, err = r.ch.NewInsert().Model(&payloads).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) filterTx(ctx context.Context, f *core.TransactionFilter) (ret []*core.Transaction, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.WithAccountState {
		q = q.Relation("Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.ExcludeColumn("code", "data") // TODO: optional
		})
		if f.WithAccountData {
			q = q.Relation("Account.StateData")
		}
	}
	if f.WithMessages {
		q = q.
			Relation("InMsg").
			Relation("InMsg.Payload").
			Relation("OutMsg").
			Relation("OutMsg.Payload")
	}

	if len(f.Hash) > 0 {
		q = q.Where("transaction.hash = ?", f.Hash)
	}
	if len(f.InMsgHash) > 0 {
		q = q.Where("transaction.in_msg_hash = ?", f.InMsgHash)
	}
	if len(f.Addresses) > 0 {
		q = q.Where("transaction.address in (?)", bun.In(f.Addresses))
	}
	if f.Workchain != nil {
		q = q.Where("transaction.block_workchain = ?", f.Workchain)
	}
	if f.BlockID != nil {
		q = q.Where("transaction.block_workchain = ?", f.BlockID.Workchain).
			Where("transaction.block_shard = ?", f.BlockID.Shard).
			Where("transaction.block_seq_no = ?", f.BlockID.SeqNo)
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where("transaction.created_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where("transaction.created_lt < ?", f.AfterTxLT)
		}
	}

	if f.Order != "" {
		q = q.Order("transaction.created_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countTx(ctx context.Context, f *core.TransactionFilter) (int, error) {
	q := r.ch.NewSelect().
		Model((*core.Transaction)(nil))

	if len(f.Hash) > 0 {
		q = q.Where("hash = ?", f.Hash)
	}
	if len(f.InMsgHash) > 0 {
		q = q.Where("in_msg_hash = ?", f.InMsgHash)
	}
	if len(f.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(f.Addresses))
	}
	if f.Workchain != nil {
		q = q.Where("block_workchain = ?", *f.Workchain)
	}
	if f.BlockID != nil {
		q = q.Where("block_workchain = ?", f.BlockID.Workchain).
			Where("block_shard = ?", f.BlockID.Shard).
			Where("block_seq_no = ?", f.BlockID.SeqNo)
	}

	return q.Count(ctx)
}

func (r *Repository) GetTransactions(ctx context.Context, f *core.TransactionFilter) (*core.TransactionFilterResults, error) {
	var (
		res = new(core.TransactionFilterResults)
		err error
	)

	res.Rows, err = r.filterTx(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countTx(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *Repository) filterMsg(ctx context.Context, f *core.MessageFilter) (ret []*core.Message, err error) {
	q := r.pg.NewSelect()
	if f.DBTx != nil {
		q = f.DBTx.NewSelect()
	}

	q = q.Model(&ret)

	if f.WithPayload {
		q = q.Relation("Payload")
	}

	if len(f.Hash) > 0 {
		q = q.Where("message.hash = ?", f.Hash)
	}
	if len(f.SrcAddresses) > 0 {
		q = q.Where("message.src_address in (?)", bun.In(f.SrcAddresses)).
			Where("length(message.src_address) > 0") // partial index
	}
	if len(f.DstAddresses) > 0 {
		q = q.Where("message.dst_address in (?)", bun.In(f.DstAddresses)).
			Where("length(message.dst_address) > 0") // partial index
	}

	if f.WithPayload {
		if len(f.SrcContracts) > 0 {
			q = q.Where("payload.src_contract IN (?)", bun.In(f.SrcContracts)).
				Where("length(payload.src_contract) > 0") // partial index
		}
		if len(f.DstContracts) > 0 {
			q = q.Where("payload.dst_contract IN (?)", bun.In(f.DstContracts)).
				Where("length(payload.dst_contract) > 0") // partial index
		}
		if len(f.OperationNames) > 0 {
			q = q.Where("payload.operation_name IN (?)", bun.In(f.OperationNames))
		}
		if f.MinterAddress != nil {
			q = q.Where("payload.minter_address = ?", f.MinterAddress)
		}
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where("message.created_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where("message.created_lt < ?", f.AfterTxLT)
		}
	}

	if f.Order != "" {
		q = q.Order("message.created_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countMsg(ctx context.Context, f *core.MessageFilter) (int, error) {
	var payload bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if f.WithPayload {
		if len(f.SrcContracts) > 0 {
			q, payload = q.Where("src_contract IN (?)", ch.In(f.SrcContracts)), true
		}
		if len(f.DstContracts) > 0 {
			q, payload = q.Where("dst_contract IN (?)", ch.In(f.DstContracts)), true
		}
		if len(f.OperationNames) > 0 {
			q, payload = q.Where("operation_name IN (?)", ch.In(f.OperationNames)), true
		}
		if f.MinterAddress != nil {
			q, payload = q.Where("minter_address = ?", f.MinterAddress), true
		}
	}

	if len(f.Hash) > 0 {
		q = q.Where("hash = ?", f.Hash)
	}
	if len(f.SrcAddresses) > 0 {
		q = q.Where("src_address in (?)", ch.In(f.SrcAddresses))
	}
	if len(f.DstAddresses) > 0 {
		q = q.Where("dst_address in (?)", ch.In(f.DstAddresses))
	}

	if payload {
		q = q.Model((*core.MessagePayload)(nil))
	} else {
		q = q.Model((*core.Message)(nil))
	}

	return q.Count(ctx)
}

func (r *Repository) GetMessages(ctx context.Context, f *core.MessageFilter) (*core.MessageFilterResults, error) {
	var (
		res = new(core.MessageFilterResults)
		err error
	)

	res.Rows, err = r.filterMsg(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countMsg(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}
