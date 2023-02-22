package contract

import (
	"context"
	"flag"
	"os"

	"github.com/allisson/go-env"
	"github.com/rs/zerolog/log"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
)

func InsertOperation() {
	var err error

	op := new(core.ContractOperation)

	f := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	name := f.String("name", "", "Unique contract operation name (example: nft_item_transfer)")
	contract := f.String("contract", "", "Contract name")
	opid := f.Uint64("opid", 0, "Operation ID")
	schema := f.String("schema", "", "Message body schema")
	_ = f.Parse(os.Args[2:])

	if *name == "" {
		log.Fatal().Msg("operation name must be set")
	}
	if *contract == "" {
		log.Fatal().Msg("contract name must be set")
	}
	if *opid == 0 {
		log.Fatal().Msg("operation id must be set")
	}
	if *schema == "" {
		log.Fatal().Msg("operation schema must be set")
	}

	op.Name = *name
	op.ContractName = abi.ContractName(*contract)
	op.OperationID = uint32(*opid)
	op.Schema = []byte(*schema)

	conn, err := repository.ConnectDB(context.Background(),
		env.GetString("DB_CH_URL", ""),
		env.GetString("DB_PG_URL", ""))
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to a database")
	}
	_, err = conn.CH.NewInsert().Model(op).Exec(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("cannot insert contract interface")
	}

	log.Info().
		Str("op_name", op.Name).
		Str("contract", string(op.ContractName)).
		Uint32("op_id", op.OperationID).
		Str("schema", string(op.Schema)).
		Msg("inserted new contract operation")
}
