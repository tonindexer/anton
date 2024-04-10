package label

import (
	"fmt"
	"io"
	"net/http"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v2"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
)

type tonscanLabel struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

func tonscanIsCexLabel(name string) bool {
	switch name {
	case "Crypto Bot", "Crypto Bot Cold Storage",
		"Wallet Bot", "Old Wallet Bot",
		"OKX", "Bitfinex", "MEXC",
		"ByBit", "ByBit Witdrawal",
		"bit.com", "Bitpapa", "FixedFloat",
		"Huobi Deposit", "Huobi Withdrawal",
		"KuCoin Deposit", "KuCoin Withdrawal",
		"FTX", "BitGo FTX Bankruptcy Custody",
		"EXMO", "EXMO Cold Storage 1", "EXMO Cold Storage 2", "EXMO Deposit",
		"CoinEx", "Gate.io",
		"AvanChange",
		"xRocket",
		"Coinone Deposit", "Coinone Withdrawal":
		return true
	default:
		return false
	}
}

func fetchTonscanLabels() ([]*core.AddressLabel, error) {
	var ret []*core.AddressLabel

	// https://raw.githubusercontent.com/menschee/tonscanplus/main/data.json

	files := []string{"community.yaml", "exchanges.yaml", "people.yaml", "scam.yaml", "system.yaml", "validators.yaml"}

	for _, f := range files {
		resp, err := http.Get("https://raw.githubusercontent.com/catchain/address-book/master/source/" + f)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("cannot fetch %s file, server returned %d code", f, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var labels []*tonscanLabel
		if err := yaml.Unmarshal(body, &labels); err != nil {
			return nil, errors.Wrapf(err, "cannot yaml unmarshal %s file", f)
		}

		for _, l := range labels {
			switch l.Name {
			case "Burn Address", "System":
				continue
			}

			a := new(addr.Address)

			err := a.UnmarshalText([]byte(l.Address))
			if err != nil {
				return nil, errors.Wrapf(err, "unmarshal %s address", l.Address)
			}

			var categories []core.LabelCategory
			if tonscanIsCexLabel(l.Name) {
				categories = append(categories, core.CentralizedExchange)
			}
			if f == "scam.yaml" {
				categories = append(categories, core.Scam)
			}

			ret = append(ret, &core.AddressLabel{
				Address:    *a,
				Name:       l.Name,
				Categories: categories,
			})
		}
	}

	return ret, nil
}

var Command = &cli.Command{
	Name:  "label",
	Usage: "Adds new address label to the database",

	ArgsUsage: "address label [category]",

	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "tonscan",
			Usage:   "add labels from tonscan",
			Aliases: []string{"c"},
		},
	},

	Action: func(ctx *cli.Context) error {
		var labels []*core.AddressLabel

		if ctx.Bool("tonscan") {
			tonscan, err := fetchTonscanLabels()
			if err != nil {
				return err
			}
			labels = append(labels, tonscan...)
		}

		if ctx.Args().Len() >= 2 {
			a := new(addr.Address)

			if err := a.UnmarshalText([]byte(ctx.Args().Get(0))); err != nil {
				return err
			}

			labels = append(labels, &core.AddressLabel{
				Address:    *a,
				Name:       ctx.Args().Get(1),
				Categories: []core.LabelCategory{core.LabelCategory(ctx.Args().Get(2))},
			})
		}

		chURL := env.GetString("DB_CH_URL", "")
		pgURL := env.GetString("DB_PG_URL", "")

		conn, err := repository.ConnectDB(ctx.Context, chURL, pgURL)
		if err != nil {
			return errors.Wrap(err, "cannot connect to a database")
		}

		accRepo := account.NewRepository(conn.CH, conn.PG)

		for _, l := range labels {
			err := accRepo.AddAddressLabel(ctx.Context, l)
			if errors.Is(err, core.ErrAlreadyExists) {
				log.Error().Err(err).Str("addr", l.Address.Base64()).Str("name", l.Name).Msg("cannot insert label")
				continue
			}
			if err != nil {
				return errors.Wrapf(err, "%s label", l.Address.String())
			}
		}

		return nil
	},
}
