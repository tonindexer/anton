package rescan

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

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/parser"
	"github.com/tonindexer/anton/internal/app/rescan"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/repository/msg"
)

var Command = &cli.Command{
	Name: "rescan",

	Usage: "Updates account states and messages data",

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
			return errors.New("no contract interfaces")
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
		i := rescan.NewService(&app.RescanConfig{
			ContractRepo: contractRepo,
			AccountRepo:  account.NewRepository(conn.CH, conn.PG),
			BlockRepo:    block.NewRepository(conn.CH, conn.PG),
			MessageRepo:  msg.NewRepository(conn.CH, conn.PG),
			Parser:       p,
			Workers:      env.GetInt("RESCAN_WORKERS", 4),
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
