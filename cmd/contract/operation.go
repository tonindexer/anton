package contract

import (
	"context"
	"database/sql"
	"flag"
	"os"

	"github.com/allisson/go-env"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

func InsertOperation() {
	op := new(core.ContractOperation)

	f := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	name := f.String("name", "", "Unique contract operation name (example: nft_item_transfer)")
	iface := f.String("contract", "", "Contract name")
	opid := f.Uint64("opid", 0, "Operation ID")
	schema := f.String("schema", "", "Message body schema")
	_ = f.Parse(os.Args[2:])

	if *name == "" {
		log.Fatal().Msg("operation name must be set")
	}
	if *iface == "" {
		log.Fatal().Msg("contract name must be set")
	}
	if *opid == 0 {
		log.Fatal().Msg("operation id must be set")
	}
	if *schema == "" {
		log.Fatal().Msg("operation schema must be set")
	}

	pg := bun.NewDB(
		sql.OpenDB(
			pgdriver.NewConnector(
				pgdriver.WithDSN(env.GetString("DB_PG_URL", "")),
			),
		),
		pgdialect.New(),
	)
	if err := pg.Ping(); err != nil {
		log.Fatal().Err(err).Msg("cannot ping postgresql")
	}

	if err := contract.NewRepository(pg).AddOperation(context.Background(), op); err != nil {
		log.Fatal().Err(err).Msg("cannot insert contract operation")
	}

	log.Info().
		Str("op_name", op.Name).
		Str("contract", string(op.ContractName)).
		Uint32("op_id", op.OperationID).
		Str("schema", string(op.Schema)).
		Msg("inserted new contract operation")
}
