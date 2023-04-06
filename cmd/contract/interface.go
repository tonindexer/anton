package contract

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/allisson/go-env"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

func InsertInterface() {
	i := new(core.ContractInterface)

	f := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	name := f.String("name", "", "Unique contract name (example: getgems_nft_sale)")
	address := f.String("address", "", "Contract address")
	code := f.String("code", "", "Contract code encoded to hex")
	getMethods := f.String("getmethods", "", "Contract get methods separated with commas")
	_ = f.Parse(os.Args[2:])

	if *name == "" {
		log.Fatal().Msg("contract name must be set")
	}
	if *address == "" && *code == "" && *getMethods == "" {
		log.Fatal().Msg("contract address or code or get methods must be set")
	}

	i.Name = abi.ContractName(*name)
	if *address != "" {
		i.Addresses = []*addr.Address{addr.MustFromBase64(*address)}
	}
	if *code != "" {
		dec, err := hex.DecodeString(*code)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot parse contract code")
		}
		codeCell, err := cell.FromBOC(dec)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot get contract code cell from boc")
		}
		i.Code = codeCell.ToBOC()
		i.CodeHash = codeCell.Hash()
	}
	if *getMethods != "" {
		i.GetMethods = strings.Split(*getMethods, ",")
	}
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
		log.Fatal().Err(err).Msg("cannot ping postgresql")
	}

	if err := contract.NewRepository(pg).AddInterface(context.Background(), i); err != nil {
		log.Fatal().Err(err).Msg("cannot insert contract interface")
	}

	log.Info().
		Str("name", string(i.Name)).
		Str("address", i.Addresses[0].Base64()).
		Str("get_methods", fmt.Sprintf("%+v", i.GetMethods)).
		Str("code", hex.EncodeToString(i.Code)).
		Msg("inserted new contract interface")
}
