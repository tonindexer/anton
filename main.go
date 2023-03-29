package main

import (
	"os"

	"github.com/allisson/go-env"

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

var availableCommands = map[string]struct {
	Description string
	Run         func()
}{
	"indexer": {
		Description: "Background task to scan new blocks",
		Run:         indexer.Run,
	},
	"query": {
		Description: "HTTP API",
		Run:         query.Run,
	},
	"archiveNodes": {
		Description: "Returns archive nodes found from config",
		Run:         archive.Run,
	},
	"addInterface": {
		Description: "Inserts new contract interface to a database",
		Run:         contract.InsertInterface,
	},
	"addOperation": {
		Description: "Inserts new contract operation to a database",
		Run:         contract.InsertOperation,
	},
}

func printHelp() {
	println("available commands:")
	for cmd, v := range availableCommands {
		println("*", cmd, "--", v.Description)
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd, ok := availableCommands[os.Args[1]]
	if !ok {
		println("[!] unknown command", os.Args[1])
		printHelp()
		os.Exit(1)
	}

	cmd.Run()
}
