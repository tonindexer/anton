package indexer

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/indexer"
	"github.com/tonindexer/anton/internal/app/parser"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/repository/tx"
)

func initDB(ctx context.Context, conn *repository.DB) error {
	abiRepo := contract.NewRepository(conn.PG)

	_, err := abiRepo.GetOperationByID(ctx, []abi.ContractName{abi.NFTItem}, false, 0x5fcc3d14)
	if err == nil {
		return nil // tables exist
	}

	log.Info().Msg("creating tables")

	err = block.CreateTables(ctx, conn.CH, conn.PG)
	if err != nil {
		return err
	}
	err = account.CreateTables(ctx, conn.CH, conn.PG)
	if err != nil {
		return err
	}
	err = tx.CreateTables(ctx, conn.CH, conn.PG)
	if err != nil {
		return err
	}
	err = msg.CreateTables(ctx, conn.CH, conn.PG)
	if err != nil {
		return err
	}
	err = contract.CreateTables(ctx, conn.PG)
	if err != nil {
		return err
	}

	if err := repository.InsertKnownInterfaces(ctx, abiRepo); err != nil {
		return errors.Wrap(err, "cannot insert interfaces")
	}

	log.Info().Msg("inserted known contract interfaces")

	return nil
}

func Run() {
	chURL := env.GetString("DB_CH_URL", "")
	pgURL := env.GetString("DB_PG_URL", "")

	conn, err := repository.ConnectDB(context.Background(), chURL, pgURL)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to a database")
	}
	if err := initDB(context.Background(), conn); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	var liteservers []*app.ServerAddr
	for _, addr := range strings.Split(env.GetString("LITESERVERS", ""), ",") {
		splitted := strings.Split(addr, "|")
		if len(splitted) != 2 {
			log.Fatal().Err(err).Msg("wrong server address format")
		}
		liteservers = append(liteservers, &app.ServerAddr{
			IPPort:    splitted[0],
			PubKeyB64: splitted[1],
		})
	}

	p, err := parser.NewService(context.Background(), &app.ParserConfig{
		DB:      conn,
		Servers: liteservers,
	})
	if err != nil {
		panic(err)
	}

	s, err := indexer.NewService(context.Background(), &app.IndexerConfig{
		DB:               conn,
		Parser:           p,
		FromBlock:        uint32(env.GetInt32("FROM_BLOCK", 22222022)),
		FetchBlockPeriod: 1 * time.Millisecond,
	})
	if err != nil {
		panic(err)
	}
	if err = s.Start(); err != nil {
		panic(err)
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
}
