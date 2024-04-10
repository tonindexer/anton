package msg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _ repository.Message = (*Repository)(nil)

type Repository struct {
	ch *ch.DB
	pg *bun.DB
}

func NewRepository(ck *ch.DB, pg *bun.DB) *Repository {
	return &Repository{ch: ck, pg: pg}
}

func createIndexes(ctx context.Context, pgDB *bun.DB) error {
	var err error

	// messages

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Column("src_address", "src_tx_lt").
		Where("src_address IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src_address source_tx_lt pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Unique().
		Column("src_address", "created_lt").
		Where("src_address IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr lt pg create unique index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("src_address").
		Where("src_address IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message src addr pg create index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("dst_address").
		Where("dst_address IS NOT NULL").
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

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("operation_id").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message operation id pg create index")
	}

	// message payloads

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("src_contract").
		Where("src_contract IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create src_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("dst_contract").
		Where("src_contract IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create dst_contract index")
	}

	_, err = pgDB.NewCreateIndex().
		Model(&core.Message{}).
		Using("HASH").
		Column("operation_name").
		Where("operation_name IS NOT NULL").
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "message payload pg create operation name index")
	}

	return nil
}

func CreateTables(ctx context.Context, chDB *ch.DB, pgDB *bun.DB) error {
	_, err := pgDB.ExecContext(ctx, "CREATE TYPE message_type AS ENUM (?, ?, ?)",
		core.ExternalIn, core.ExternalOut, core.Internal)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "messages pg create enum")
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

	_, err = pgDB.ExecContext(ctx, `
ALTER TABLE messages
ADD CONSTRAINT messages_tx_lt_notnull
CHECK (
    (type = 'EXTERNAL_OUT' AND src_address IS NOT NULL AND src_tx_lt IS NOT NULL AND dst_address IS NULL AND dst_tx_lt IS NULL) OR
    (type = 'EXTERNAL_IN' AND src_address IS NULL AND src_tx_lt IS NULL AND dst_address IS NOT NULL AND dst_tx_lt IS NOT NULL) OR
    (type = 'INTERNAL' AND (src_workchain != -1 OR dst_workchain != -1) AND src_tx_lt IS NOT NULL AND dst_tx_lt IS NOT NULL) OR
    (type = 'INTERNAL' AND src_workchain = -1 AND dst_workchain = -1)
)`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "messages pg create source tx hash check")
	}

	if err := createIndexes(ctx, pgDB); err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddMessages(ctx context.Context, tx bun.Tx, messages []*core.Message) error {
	if len(messages) == 0 {
		return nil
	}
	for _, msg := range messages { // TODO: on conflict does not work with array (bun bug)
		// some external messages can be repeated with the same hash

		// if some message has been already inserted,
		// we update destination transaction and parsed data

		_, err := tx.NewInsert().Model(msg).
			On("CONFLICT (hash) DO UPDATE").
			Set("dst_tx_lt = ?dst_tx_lt").
			Set("dst_workchain = ?dst_workchain").
			Set("dst_shard = ?dst_shard").
			Set("dst_block_seq_no = ?dst_block_seq_no").
			Set("src_contract = ?src_contract").
			Set("dst_contract = ?dst_contract").
			Set("operation_name = ?operation_name").
			Set("data_json = ?data_json").
			Set("error = ?error").
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	_, err := r.ch.NewInsert().Model(&messages).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateMessages(ctx context.Context, messages []*core.Message) error {
	if len(messages) == 0 {
		return nil
	}

	for _, msg := range messages {
		log.Debug().
			Hex("msg_hash", msg.Hash).
			Str("src_address", msg.SrcAddress.Base64()).
			Str("src_contract", string(msg.SrcContract)).
			Str("dst_address", msg.DstAddress.Base64()).
			Str("dst_contract", string(msg.DstContract)).
			Uint32("operation_id", msg.OperationID).
			Str("operation_name", msg.OperationName).
			RawJSON("data_json", msg.DataJSON).
			Str("error", msg.Error).
			Msg("updating message")

		_, err := r.pg.NewUpdate().Model(msg).
			Set("src_contract = ?src_contract").
			Set("dst_contract = ?dst_contract").
			Set("operation_name = ?operation_name").
			Set("data_json = ?data_json").
			Set("error = ?error").
			WherePK().
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	_, err := r.ch.NewInsert().Model(&messages).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetMessage(ctx context.Context, hash []byte) (*core.Message, error) {
	var ret core.Message

	err := r.pg.NewSelect().Model(&ret).
		Relation("SrcState").
		Relation("DstState").
		Where("hash = ?", hash).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (r *Repository) GetMessages(ctx context.Context, hashes [][]byte) ([]*core.Message, error) {
	var ret []*core.Message

	err := r.pg.NewSelect().Model(&ret).
		Where("hash IN (?)", bun.In(hashes)).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// MatchMessagesByOperationDesc returns hashes of suitable messages for the given contract operation.
func (r *Repository) MatchMessagesByOperationDesc(ctx context.Context,
	contractName abi.ContractName,
	msgType core.MessageType,
	outgoing bool,
	operationId uint32,
	afterAddress *addr.Address,
	afterTxLt uint64,
	limit int,
) ([][]byte, error) {
	var addressesRet []struct {
		Address *addr.Address `ch:"type:String"`
	}

	q := r.ch.NewSelect().Model((*core.AccountState)(nil)).
		ColumnExpr("DISTINCT address").
		Where("hasAny(types, [?])", string(contractName))
	if afterAddress != nil {
		q = q.Where("address >= ?", afterAddress)
	}
	err := q.
		Order("address ASC").
		Limit(limit).
		Scan(ctx, &addressesRet)
	if err != nil {
		return nil, errors.Wrap(err, "get contract addresses")
	}

	var addresses []*addr.Address
	for _, row := range addressesRet {
		addresses = append(addresses, row.Address)
	}

	addrCol, ltCol := "dst_address", "dst_tx_lt"
	if outgoing {
		addrCol, ltCol = "src_address", "src_tx_lt"
	}

	var msgHashesRet []struct {
		Hash []byte
	}
	q = r.ch.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("DISTINCT hash").
		Where("type = ?", string(msgType)).
		Where(addrCol+" IN (?)", ch.In(addresses)).
		Where("operation_id = ?", operationId)
	if afterAddress != nil && afterTxLt != 0 {
		q = q.Where(fmt.Sprintf("(%s, %s) > (?, ?)", addrCol, ltCol), afterAddress, afterTxLt)
	}
	err = q.
		OrderExpr(fmt.Sprintf("%s ASC, %s ASC", addrCol, ltCol)).
		Limit(limit).
		Scan(ctx, &msgHashesRet)
	if err != nil {
		return nil, errors.Wrap(err, "get message hashes")
	}

	var hashes [][]byte
	for _, row := range msgHashesRet {
		hashes = append(hashes, row.Hash)
	}
	return hashes, nil
}
