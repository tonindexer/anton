package archive

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func intToIP4(ipInt int64) string {
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt(ipInt&0xff, 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

var Command = &cli.Command{
	Name:  "archive",
	Usage: "Prints archive nodes found from config",

	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "testnet",
			Aliases: []string{"t"},
			Value:   false,
			Usage:   "use testnet",
		},
	},

	Action: func(ctx *cli.Context) error {
		url := "https://ton-blockchain.github.io/global.config.json"
		if ctx.Bool("testnet") {
			url = "https://ton-blockchain.github.io/testnet-global.config.json"
		}

		cfg, err := liteclient.GetConfigFromUrl(ctx.Context, url)
		if err != nil {
			return err
		}

		for i := range cfg.Liteservers {
			ls := &cfg.Liteservers[i]

			client := liteclient.NewConnectionPool()

			addr := fmt.Sprintf("%s:%d", intToIP4(ls.IP), ls.Port)
			if err := client.AddConnection(ctx.Context, addr, ls.ID.Key); err != nil {
				continue
			}

			api := ton.NewAPIClient(client)

			master, err := api.GetMasterchainInfo(ctx.Context)
			if err != nil {
				continue
			}

			_, err = api.LookupBlock(ctx.Context, master.Workchain, master.Shard, 3)
			if err != nil {
				continue
			}

			log.Info().Str("addr", addr).Str("key", ls.ID.Key).Msg("new archive liteserver")
		}

		return nil
	},
}
