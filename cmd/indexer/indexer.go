package indexer

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/internal/app"
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
		var liteservers []*app.ServerAddr

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
			return errors.Wrap(err, "no contract interfaces")
		}

		for _, addr := range strings.Split(env.GetString("LITESERVERS", ""), ",") {
			split := strings.Split(addr, "|")
			if len(split) != 2 {
				return errors.Wrapf(err, "wrong server address format '%s'", addr)
			}
			liteservers = append(liteservers, &app.ServerAddr{
				IPPort:    split[0],
				PubKeyB64: split[1],
			})
		}

		p, err := parser.NewService(ctx.Context, &app.ParserConfig{
			DB:      conn,
			Servers: liteservers,
		})
		if err != nil {
			return errors.Wrap(err, "new parser service")
		}

		s, err := indexer.NewService(ctx.Context, &app.IndexerConfig{
			DB:               conn,
			Parser:           p,
			FromBlock:        uint32(env.GetInt32("FROM_BLOCK", 22222022)),
			FetchBlockPeriod: 1 * time.Millisecond,
		})
		if err != nil {
			return err
		}
		if err = s.Start(); err != nil {
			return err
		}

		c := make(chan os.Signal, 1)
		done := make(chan struct{}, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			s.Stop()
			done <- struct{}{}
		}()

		<-done

		return nil
	},
}
