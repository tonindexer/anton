package main

import (
	"fmt"
	"os"

	"github.com/allisson/go-env"
	"github.com/urfave/cli/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/cmd/archive"
	"github.com/tonindexer/anton/cmd/contract"
	"github.com/tonindexer/anton/cmd/indexer"
	"github.com/tonindexer/anton/cmd/query"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	level := zerolog.InfoLevel
	if env.GetBool("DEBUG_LOGS", false) {
		level = zerolog.DebugLevel
	}

	// add file and line number to log
	log.Logger = log.With().Caller().Logger().Level(level)
}

func main() {
	app := &cli.App{
		Name:  "anton",
		Usage: "an indexing project",
		Commands: []*cli.Command{
			indexer.Command,
			query.Command,
			archive.Command,
			contract.InterfaceCommand,
			contract.OperationCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
