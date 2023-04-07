package query

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/allisson/go-env"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/api/http"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/query"
	"github.com/tonindexer/anton/internal/core/repository"
)

func Run() {
	chURL := env.GetString("DB_CH_URL", "")
	pgURL := env.GetString("DB_PG_URL", "")

	conn, err := repository.ConnectDB(context.Background(), chURL, pgURL)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to a database")
	}

	qs, err := query.NewService(context.Background(), &app.QueryConfig{
		DB: conn,
	})
	if err != nil {
		panic(err)
	}

	srv := http.NewServer(
		env.GetString("LISTEN", "0.0.0.0:80"),
	)
	srv.RegisterRoutes(http.NewController(qs))

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		conn.Close()
		os.Exit(0)
	}()

	if err = srv.Run(); err != nil {
		panic(err)
	}
}
