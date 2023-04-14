package contract

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

var InterfaceCommand = &cli.Command{
	Name:     "addInterface",
	Usage:    "Adds contract interface to the database",
	Category: "abi",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "contract",
			Required: true,
			Aliases:  []string{"n"},
			Usage:    "Unique contract name (example: getgems_nft_sale)",
		},
		&cli.StringSliceFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "Contract addresses",
		},
		&cli.StringFlag{
			Name:    "code",
			Aliases: []string{"c"},
			Usage:   "Contract code BoC encoded to hex",
		},
		&cli.StringSliceFlag{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Contract get methods",
		},
	},

	Action: func(ctx *cli.Context) error {
		var i core.ContractInterface

		addresses := ctx.StringSlice("address")
		codeStr := ctx.String("code")
		getMethods := ctx.StringSlice("get")

		if addresses == nil && codeStr == "" && getMethods == nil {
			return errors.New("contract addresses or code or get methods must be set")
		}

		i.Name = abi.ContractName(ctx.String("contract"))

		for _, addrStr := range addresses {
			a, err := new(addr.Address).FromString(addrStr)
			if err != nil {
				return errors.Wrapf(err, "parse %s", addrStr)
			}
			i.Addresses = append(i.Addresses, a)
		}

		if codeStr != "" {
			dec, err := hex.DecodeString(codeStr)
			if err != nil {
				return errors.Wrapf(err, "cannot parse contract code")
			}
			codeCell, err := cell.FromBOC(dec)
			if err != nil {
				return errors.Wrapf(err, "cannot get contract code cell from boc")
			}
			i.Code = codeCell.ToBOC()
			i.CodeHash = codeCell.Hash()
		}

		i.GetMethods = getMethods
		for _, get := range i.GetMethods {
			i.GetMethodHashes = append(i.GetMethodHashes, abi.MethodNameHash(get))
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
			return errors.Wrapf(err, "cannot ping postgresql")
		}

		if err := contract.NewRepository(pg).AddInterface(ctx.Context, &i); err != nil {
			return errors.Wrapf(err, "cannot insert contract interface")
		}

		log.Info().
			Str("name", string(i.Name)).
			Str("address", i.Addresses[0].Base64()).
			Str("code", hex.EncodeToString(i.Code)).
			Str("get_methods", fmt.Sprintf("%+v", i.GetMethods)).
			Str("get_method_hashes", fmt.Sprintf("%+v", i.GetMethodHashes)).
			Msg("inserted new contract interface")

		return nil
	},
}
