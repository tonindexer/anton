package contract

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

var OperationCommand = &cli.Command{
	Name:     "addOperation",
	Usage:    "Adds contract operation to the database",
	Category: "abi",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "contract",
			Required: true,
			Aliases:  []string{"n"},
			Usage:    "Contract name (example: getgems_nft_sale)",
		},
		&cli.StringFlag{
			Name:     "operation",
			Required: true,
			Aliases:  []string{"op"},
			Usage:    "Unique contract operation name (example: nft_item_transfer)",
		},
		&cli.Uint64Flag{
			Name:     "operationId",
			Required: true,
			Aliases:  []string{"id"},
			Usage:    "Contract addresses",
		},
		&cli.BoolFlag{
			Name:    "outgoing",
			Aliases: []string{"o"},
			Usage:   "Does the message go from the given contract",
			Value:   false,
		},
		&cli.StringFlag{
			Name:     "schema",
			Required: true,
			Aliases:  []string{"s"},
			Usage:    "Message body schema",
		},
	},

	Action: func(ctx *cli.Context) error {
		var op core.ContractOperation

		op.Name = ctx.String("operation")
		op.ContractName = abi.ContractName(ctx.String("contract"))
		op.Outgoing = ctx.Bool("outgoing")
		op.OperationID = uint32(ctx.Uint64("operationId"))

		schema := json.RawMessage(ctx.String("schema"))
		if !json.Valid(schema) {
			return fmt.Errorf("json is not valid: %s", string(schema))
		}
		op.Schema = schema

		pg := bun.NewDB(
			sql.OpenDB(
				pgdriver.NewConnector(
					pgdriver.WithDSN(env.GetString("DB_PG_URL", "")),
				),
			),
			pgdialect.New(),
		)
		if err := pg.Ping(); err != nil {
			return errors.Wrap(err, "cannot ping postgresql")
		}

		if err := contract.NewRepository(pg).AddOperation(ctx.Context, &op); err != nil {
			return errors.Wrap(err, "cannot insert contract operation")
		}

		log.Info().
			Str("op_name", op.Name).
			Str("contract", string(op.ContractName)).
			Uint32("op_id", op.OperationID).
			Str("schema", string(op.Schema)).
			Msg("inserted new contract operation")

		return nil
	},
}
