package contract

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/allisson/go-env"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/cli/v2"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

func readContractInterfaces(filenames []string) (ret []*abi.InterfaceDesc, err error) {
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

	return &core.ContractOperation{
		Name:         d.Name,
		ContractName: t,
		Outgoing:     false,
		OperationID:  opId,
		Schema:       *d,
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

func parseInterfacesDesc(descriptors []*abi.InterfaceDesc) (retI []*core.ContractInterface, retOp []*core.ContractOperation, err error) {
	for _, d := range descriptors {
		i, operations, err := parseInterfaceDesc(d)
		if err != nil {
			return nil, nil, err
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

	Action: func(ctx *cli.Context) error {
		filenames := ctx.Args().Slice()

		interfacesDesc, err := readContractInterfaces(filenames)
		if err != nil {
			return err
		}

		interfaces, operations, err := parseInterfacesDesc(interfacesDesc)
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

		for _, i := range interfaces {
			if err := contract.NewRepository(pg).AddInterface(ctx.Context, i); err != nil {
				return errors.Wrapf(err, "cannot insert %s contract interface", i.Name)
			}
		}
		for _, op := range operations {
			if err := contract.NewRepository(pg).AddOperation(ctx.Context, op); err != nil {
				return errors.Wrapf(err, "cannot insert %s %s contract operation", op.ContractName, op.Name)
			}
		}

		return nil
	},
}
