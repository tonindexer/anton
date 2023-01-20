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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/app/indexer"
	"github.com/iam047801/tonidx/internal/app/parser"
	"github.com/iam047801/tonidx/internal/core/db"
	"github.com/iam047801/tonidx/internal/core/repository/account"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// add file and line number to log
	log.Logger = log.With().Caller().Logger().Level(zerolog.InfoLevel)
}

func initDB(ctx context.Context, conn *ch.DB) error {
	ifaces, err := account.NewRepository(conn).GetContractInterfaces(ctx)
	if err == nil && len(ifaces) > 0 {
		for _, iface := range ifaces {
			log.Debug().Str("addr", iface.Address).Str("name", string(iface.Name)).Msg("found contract interface")
		}
		return nil
	}

	log.Info().Msg("creating tables")
	if err := db.CreateTables(ctx, conn); err != nil {
		return errors.Wrap(err, "cannot create tables")
	}

	log.Info().Msg("inserting known contract interfaces")
	if err := db.InsertKnownInterfaces(ctx, conn); err != nil {
		return errors.Wrap(err, "cannot insert interfaces")
	}

	return nil
}

func Run() {
	dbURL := env.GetString("DB_URL", "")

	conn, err := db.Connect(context.Background(), dbURL)
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
		FetchBlockPeriod: 10 * time.Millisecond,
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
