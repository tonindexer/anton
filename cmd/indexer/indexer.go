package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/abi"
	contractDesc "github.com/tonindexer/anton/cmd/contract"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/fetcher"
	"github.com/tonindexer/anton/internal/app/indexer"
	"github.com/tonindexer/anton/internal/app/parser"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

func getAllKnownContractFilenames(contractsDir string) (res []string, err error) {
	entries, err := os.ReadDir(contractsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "read %s directory", contractsDir)
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			res = append(res, e.Name())
		}
	}
	return res, nil
}

func addKnownContracts(ctx context.Context, pg *bun.DB, contractsDir string) error {
	filenames, err := getAllKnownContractFilenames(contractsDir)
	if err != nil {
		return err
	}

	if len(filenames) == 0 {
		return fmt.Errorf("json files are not found inside contracts directory")
	}

	tx, err := pg.Begin()
	if err != nil {
		return errors.Wrap(err, "begin postgresql transaction")
	}
	defer func() { _ = tx.Rollback() }()

	for _, fn := range filenames {
		var descriptions []*abi.InterfaceDesc

		j, err := os.ReadFile(contractsDir + "/" + fn)
		if err != nil {
			return errors.Wrapf(err, "read %s", fn)
		}

		if err := json.Unmarshal(j, &descriptions); err != nil {
			return errors.Wrapf(err, "unmarshal json")
		}

		definitions, interfaces, operations, err := contractDesc.ParseInterfacesDesc(descriptions)
		if err != nil {
			return err
		}

		for dn, d := range definitions {
			def := &core.ContractDefinition{
				Name:   dn,
				Schema: d,
			}
			_, err := tx.NewInsert().Model(def).Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "cannot insert %s definition from %s", dn, filenames)
			}
		}
		for _, i := range interfaces {
			_, err := tx.NewInsert().Model(i).Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "cannot insert %s interface from %s", i.Name, filenames)
			}
		}
		for _, op := range operations {
			_, err := tx.NewInsert().Model(op).Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "cannot insert %s operation from %s", op.OperationName, filenames)
			}
		}

		log.Info().Str("filename", fn).Msg("processed new contracts description")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit postgresql transaction")
	}

	return nil
}

var Command = &cli.Command{
	Name:    "indexer",
	Aliases: []string{"idx"},
	Usage:   "Scans new blocks",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "contracts-dir",
			Usage:   "the directory containing known contract descriptions that are added during the initial startup",
			Value:   "/var/anton/known",
			Aliases: []string{"c"},
		},
	},

	Action: func(ctx *cli.Context) error {
		chURL := env.GetString("DB_CH_URL", "")
		pgURL := env.GetString("DB_PG_URL", "")

		conn, err := repository.ConnectDB(ctx.Context, chURL, pgURL)
		if err != nil {
			return errors.Wrap(err, "cannot connect to a database")
		}

		contractRepo := contract.NewRepository(conn.PG)

		interfaces, err := contractRepo.GetInterfaces(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "get interfaces")
		}
		if len(interfaces) == 0 {
			log.Info().Str("contracts_directory", ctx.String("contracts-dir")).
				Msg("contract interfaces are not detected, inserting descriptions for known contracts")
			if err := addKnownContracts(ctx.Context, conn.PG, ctx.String("contracts-dir")); err != nil {
				return err
			}
		}

		def, err := contractRepo.GetDefinitions(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "get definitions")
		}
		err = abi.RegisterDefinitions(def)
		if err != nil {
			return errors.Wrap(err, "get definitions")
		}

		client := liteclient.NewConnectionPool()
		api := ton.NewAPIClient(client, ton.ProofCheckPolicyUnsafe).WithRetry()
		for _, addr := range strings.Split(env.GetString("LITESERVERS", ""), ",") {
			split := strings.Split(addr, "|")
			if len(split) != 2 {
				return fmt.Errorf("wrong server address format '%s'", addr)
			}
			host, key := split[0], split[1]
			if err := client.AddConnection(ctx.Context, host, key); err != nil {
				return errors.Wrapf(err, "cannot add connection with %s host and %s key", host, key)
			}
		}
		bcConfig, err := app.GetBlockchainConfig(ctx.Context, api)
		if err != nil {
			return errors.Wrap(err, "cannot get blockchain config")
		}

		p := parser.NewService(&app.ParserConfig{
			BlockchainConfig: bcConfig,
			ContractRepo:     contractRepo,
		})
		f := fetcher.NewService(&app.FetcherConfig{
			API:         api,
			AccountRepo: account.NewRepository(conn.CH, conn.PG),
			Parser:      p,
		})
		i := indexer.NewService(&app.IndexerConfig{
			DB:        conn,
			API:       api,
			Parser:    p,
			Fetcher:   f,
			FromBlock: uint32(env.GetInt32("FROM_BLOCK", 1)),
			Workers:   env.GetInt("WORKERS", 4),
		})
		if err = i.Start(); err != nil {
			return err
		}

		c := make(chan os.Signal, 1)
		done := make(chan struct{}, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			i.Stop()
			conn.Close()
			done <- struct{}{}
		}()

		<-done

		return nil
	},
}
