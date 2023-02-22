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

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/app/indexer"
	"github.com/iam047801/tonidx/internal/app/parser"
	"github.com/iam047801/tonidx/internal/core/repository"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
)

func initDB(ctx context.Context, conn *repository.DB) error {
	_, err := contract.NewRepository(conn.PG).
		GetOperationByID(ctx,
			[]abi.ContractName{abi.NFTItem}, false, 0x5fcc3d14)
	if err == nil {
		return nil // tables exist
	}

	log.Info().Msg("creating tables")
	if err := repository.CreateTablesDB(ctx, conn); err != nil {
		return errors.Wrap(err, "cannot create tables")
	}

	log.Info().Msg("inserting known contract interfaces")
	if err := repository.InsertKnownInterfaces(ctx, conn.PG); err != nil {
		return errors.Wrap(err, "cannot insert interfaces")
	}

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

	c := make(chan os.Signal)
	done := make(chan struct{})
	//nolint
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		s.Stop()
		done <- struct{}{}
	}()

	<-done
}
