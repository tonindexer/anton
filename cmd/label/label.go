package label

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
)

type tonscanLabel struct {
	TonIcon string `json:"tonIcon"`
	Name    string `json:"name"`
	IsScam  bool   `json:"isScam"`
}

func isCEX(name string) bool {
	switch name {
	case "CryptoBot", "CryptoBot Cold Storage",
		"Wallet Bot", "Old Wallet Bot",
		"OKX", "Bitfinex", "MEXC",
		"ByBit", "ByBit Witdrawal",
		"bit.com", "Bitpapa", "FixedFloat",
		"Huobi Deposit", "Huobi Withdrawal",
		"KuCoin Deposit", "KuCoin Withdrawal",
		"FTX",
		"BitGo FTX Bankruptcy Custody",
		"EXMO", "EXMO Cold Storage 1", "EXMO Cold Storage 2", "EXMO Deposit",
		"CoinEx", "Gate.io":
		return true
	default:
		return false
	}
}

func unmarshalTonscanLabel(addrStr string, j json.RawMessage) (*core.AddressLabel, error) {
	var l tonscanLabel
	var ret core.AddressLabel

	a := new(addr.Address)

	err := a.UnmarshalText([]byte(addrStr))
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal %s address", addrStr)
	}

	ret.Address = *a

	if j[0] == '"' {
		ret.Name = string(j[1 : len(j)-1])
	} else {
		err := json.Unmarshal(j, &l)
		if err != nil {
			return nil, err
		}
		ret.Name = l.Name
		if l.IsScam {
			ret.Categories = append(ret.Categories, core.Scam)
		}
	}
	if isCEX(ret.Name) {
		ret.Categories = append(ret.Categories, core.CentralizedExchange)
	}

	return &ret, nil
}

func FetchTonscanLabels() ([]*core.AddressLabel, error) {
	var ret []*core.AddressLabel

	// https://raw.githubusercontent.com/menschee/tonscanplus/main/data.json

	res, err := http.Get("https://raw.githubusercontent.com/catchain/tonscan/master/src/addrbook.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var addrMap = make(map[string]json.RawMessage)
	if err := json.Unmarshal(body, &addrMap); err != nil {
		return nil, errors.Wrap(err, "tonscan data unmarshal")
	}

	for a, j := range addrMap {
		l, err := unmarshalTonscanLabel(a, j)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal %s label: %s", a, string(j))
		}
		if l.Name == "Burn Address" {
			continue
		}
		ret = append(ret, l)
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
			tonscan, err := FetchTonscanLabels()
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
			if err != nil {
				return errors.Wrapf(err, "%s label", l.Address.String())
			}
		}

		return nil
	},
}
