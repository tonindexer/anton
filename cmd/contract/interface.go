package contract

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

func readStdin() ([]*abi.InterfaceDesc, error) {
	var interfaces []*abi.InterfaceDesc

	j, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(j, &interfaces); err != nil {
		return nil, errors.Wrapf(err, "unmarshal json")
	}

	return interfaces, nil
}

func readFiles(filenames []string) (ret []*abi.InterfaceDesc, err error) {
	for _, fn := range filenames {
		var interfaces []*abi.InterfaceDesc

		j, err := os.ReadFile(fn)
		if err != nil {
			return nil, errors.Wrapf(err, "read %s", fn)
		}

		if err := json.Unmarshal(j, &interfaces); err != nil {
			return nil, errors.Wrapf(err, "unmarshal json")
		}

		ret = append(ret, interfaces...)
	}

	return
}

func parseOperationDesc(t abi.ContractName, d *abi.OperationDesc) (*core.ContractOperation, error) {
	var opId uint32

	if c := d.Code; strings.HasPrefix(c, "0x") {
		n := new(big.Int)
		_, ok := n.SetString(c[2:], 16)
		if !ok {
			return nil, fmt.Errorf("wrong hex %s operation id format: %s", d.Name, d.Code)
		}
		opId = uint32(n.Uint64())
	} else {
		n, err := strconv.ParseUint(c, 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, "parse %s operation id", d.Name)
		}
		opId = uint32(n)
	}

	// this is needed to map interface definitions into schema
	x, err := d.New()
	if err != nil {
		return nil, errors.Wrapf(err, "creating new operation structure")
	}
	_, err = abi.NewOperationDesc(x)
	if err != nil {
		return nil, errors.Wrapf(err, "creating new operation descriptor")
	}

	if d.Type == "" {
		d.Type = string(core.Internal)
	}

	return &core.ContractOperation{
		OperationName: d.Name,
		ContractName:  t,
		MessageType:   core.MessageType(strings.ToUpper(d.Type)),
		Outgoing:      false,
		OperationID:   opId,
		Schema:        *d,
	}, nil
}

func parseInterfaceDesc(d *abi.InterfaceDesc) (*core.ContractInterface, []*core.ContractOperation, error) {
	var operations []*core.ContractOperation

	code, err := base64.StdEncoding.DecodeString(d.CodeBoc)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "decode code boc from base64")
	}

	i := core.ContractInterface{
		Name:           d.Name,
		Addresses:      d.Addresses,
		Code:           code,
		GetMethodsDesc: d.GetMethods,
	}
	for it := range i.GetMethodsDesc {
		i.GetMethodHashes = append(i.GetMethodHashes, abi.MethodNameHash(i.GetMethodsDesc[it].Name))
	}
	if len(i.Code) == 0 {
		i.Code = nil
	}

	for it := range d.InMessages {
		op, err := parseOperationDesc(i.Name, &d.InMessages[it])
		if err != nil {
			return nil, nil, err
		}
		op.Outgoing = false
		operations = append(operations, op)
	}

	for it := range d.OutMessages {
		op, err := parseOperationDesc(i.Name, &d.OutMessages[it])
		if err != nil {
			return nil, nil, err
		}
		op.Outgoing = true
		operations = append(operations, op)
	}

	return &i, operations, nil
}

func parseInterfacesDesc(descriptors []*abi.InterfaceDesc) (retD map[abi.TLBType]abi.TLBFieldsDesc, retI []*core.ContractInterface, retOp []*core.ContractOperation, _ error) {
	retD = map[abi.TLBType]abi.TLBFieldsDesc{}
	for _, desc := range descriptors {
		err := abi.RegisterDefinitions(desc.Definitions)
		if err != nil {
			return nil, nil, nil, err
		}
		for dn, d := range desc.Definitions {
			retD[dn] = d
		}
	}
	for _, desc := range descriptors {
		i, operations, err := parseInterfaceDesc(desc)
		if err != nil {
			return nil, nil, nil, err
		}
		retI = append(retI, i)
		retOp = append(retOp, operations...)
	}
	return
}

var Command = &cli.Command{
	Name:  "contract",
	Usage: "Adds contract interface to the database",

	ArgsUsage: "[file1.json] [file2.json]",

	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "stdin",
			Usage:   "read from stdin instead of files",
			Aliases: []string{"i"},
		},
	},

	Subcommands: cli.Commands{
		{
			Name:  "delete",
			Usage: "Deletes contract interface from the database",

			ArgsUsage: "[interface_name_1] [interface_name_2]",

			Action: func(ctx *cli.Context) error {
				pg := bun.NewDB(
					sql.OpenDB(
						pgdriver.NewConnector(
							pgdriver.WithDSN(env.GetString("DB_PG_URL", "")),
						),
					),
					pgdialect.New(),
				)
				if err := pg.Ping(); err != nil {
					return errors.Wrapf(err, "cannot ping postgresql")
				}

				contractRepo := contract.NewRepository(pg)

				for _, i := range ctx.Args().Slice() {
					err := contractRepo.DelInterface(ctx.Context, i)
					if err != nil {
						return errors.Wrapf(err, "deleting %s interface", i)
					}
				}

				return nil
			},
		},
	},

	Action: func(ctx *cli.Context) (err error) {
		var interfacesDesc []*abi.InterfaceDesc

		if ctx.Bool("stdin") {
			interfacesDesc, err = readStdin()
		} else {
			filenames := ctx.Args().Slice()
			if len(filenames) == 0 {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}
			interfacesDesc, err = readFiles(filenames)
		}
		if err != nil {
			return err
		}

		definitions, interfaces, operations, err := parseInterfacesDesc(interfacesDesc)
		if err != nil {
			return err
		}

		pg := bun.NewDB(
			sql.OpenDB(
				pgdriver.NewConnector(
					pgdriver.WithDSN(env.GetString("DB_PG_URL", "")),
				),
			),
			pgdialect.New(),
		)
		if err := pg.Ping(); err != nil {
			return errors.Wrapf(err, "cannot ping postgresql")
		}

		for dn, d := range definitions {
			if err := contract.NewRepository(pg).AddDefinition(ctx.Context, dn, d); err != nil {
				log.Err(err).Str("definition_name", string(dn)).Msg("cannot insert contract interface")
			}
		}
		for _, i := range interfaces {
			if err := contract.NewRepository(pg).AddInterface(ctx.Context, i); err != nil {
				log.Err(err).Str("interface_name", string(i.Name)).Msg("cannot insert contract interface")
			}
		}
		for _, op := range operations {
			if err := contract.NewRepository(pg).AddOperation(ctx.Context, op); err != nil {
				log.Err(err).
					Str("interface_name", string(op.ContractName)).
					Str("operation_name", op.OperationName).
					Msg("cannot insert contract operation")
			}
		}

		return nil
	},
}
