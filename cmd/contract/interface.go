package contract

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/allisson/go-env"
	"github.com/rs/zerolog/log"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
)

func InsertInterface() {
	var err error

	contract := new(core.ContractInterface)

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

	contract.Name = abi.ContractName(*name)
	if *address != "" {
		contract.Addresses = []*addr.Address{addr.MustFromBase64(*address)}
	}
	if *code != "" {
		contract.Code, err = hex.DecodeString(*code)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot parse contract code")
		}
	}
	if *getMethods != "" {
		contract.GetMethods = strings.Split(*getMethods, ",")
	}
	for _, get := range contract.GetMethods {
		contract.GetMethodHashes = append(contract.GetMethodHashes, abi.MethodNameHash(get))
	}

	conn, err := repository.ConnectDB(context.Background(),
		env.GetString("DB_CH_URL", ""),
		env.GetString("DB_PG_URL", ""))
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to a database")
	}
	_, err = conn.CH.NewInsert().Model(contract).Exec(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("cannot insert contract interface")
	}

	log.Info().
		Str("name", string(contract.Name)).
		Str("address", contract.Addresses[0].Base64()).
		Str("get_methods", fmt.Sprintf("%+v", contract.GetMethods)).
		Str("code", hex.EncodeToString(contract.Code)).
		Msg("inserted new contract interface")
}
