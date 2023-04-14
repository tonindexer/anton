package web

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/internal/api/http"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/query"
	"github.com/tonindexer/anton/internal/core/repository"
)

var Command = &cli.Command{
	Name:  "web",
	Usage: "HTTP JSON API",

	Action: func(ctx *cli.Context) error {
		chURL := env.GetString("DB_CH_URL", "")
		pgURL := env.GetString("DB_PG_URL", "")

		conn, err := repository.ConnectDB(
			ctx.Context, chURL, pgURL)
		if err != nil {
			return errors.Wrap(err, "cannot connect to a database")
		}

		qs, err := query.NewService(ctx.Context, &app.QueryConfig{
			DB: conn,
		})
		if err != nil {
			return err
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
			return err
		}

		return nil
	},
}
