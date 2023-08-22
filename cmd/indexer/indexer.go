package indexer

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/fetcher"
	"github.com/tonindexer/anton/internal/app/indexer"
	"github.com/tonindexer/anton/internal/app/parser"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

var Command = &cli.Command{
	Name:    "indexer",
	Aliases: []string{"idx"},
	Usage:   "Scans new blocks",

	Action: func(ctx *cli.Context) error {
		chURL := env.GetString("DB_CH_URL", "")
		pgURL := env.GetString("DB_PG_URL", "")

		conn, err := repository.ConnectDB(ctx.Context, chURL, pgURL)
		if err != nil {
			return errors.Wrap(err, "cannot connect to a database")
		}

		interfaces, err := contract.NewRepository(conn.PG).GetInterfaces(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "get interfaces")
		}
		if len(interfaces) == 0 {
			return errors.New("no contract interfaces")
		}

		client := liteclient.NewConnectionPool()
		api := ton.NewAPIClient(client)
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
			ContractRepo:     contract.NewRepository(conn.PG),
		})
		f := fetcher.NewService(&app.FetcherConfig{
			API:    api,
			Parser: p,
		})
		i := indexer.NewService(&app.IndexerConfig{
			DB:               conn,
			API:              api,
			Parser:           p,
			Fetcher:          f,
			FromBlock:        uint32(env.GetInt32("FROM_BLOCK", 1)),
			Workers:          env.GetInt("WORKERS", 4),
			InsertBlockBatch: env.GetInt("INSERT_BLOCK_BATCH", 10),
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
