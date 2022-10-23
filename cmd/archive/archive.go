package archive

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func intToIP4(ipInt int64) string {
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

func Run() {
	ctx := context.Background()

	f := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	testnet := f.Bool("testnet", false, "Using testnet")
	_ = f.Parse(os.Args[2:])

	url := "https://ton-blockchain.github.io/global.config.json"
	if *testnet {
		url = "https://ton-blockchain.github.io/testnet-global.config.json"
	}

	cfg, err := liteclient.GetConfigFromUrl(ctx, url)
	if err != nil {
		panic(err)
	}

	for i := range cfg.Liteservers {
		ls := &cfg.Liteservers[i]

		client := liteclient.NewConnectionPool()

		addr := fmt.Sprintf("%s:%d", intToIP4(ls.IP), ls.Port)
		if err := client.AddConnection(ctx, addr, ls.ID.Key); err != nil {
			continue
		}

		api := ton.NewAPIClient(client)

		master, err := api.GetMasterchainInfo(ctx)
		if err != nil {
			continue
		}

		_, err = api.LookupBlock(ctx, master.Workchain, master.Shard, 3)
		if err != nil {
			continue
		}

		log.Info().Str("addr", addr).Str("key", ls.ID.Key).Msg("new archive liteserver")
	}
}
