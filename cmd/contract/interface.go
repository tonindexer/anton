package contract

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
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
	"github.com/tonindexer/anton/internal/core/repository/rescan"
)

func dbConnect() (*bun.DB, error) {
	pg := bun.NewDB(
		sql.OpenDB(
			pgdriver.NewConnector(
				pgdriver.WithDSN(env.GetString("DB_PG_URL", "")),
			),
		),
		pgdialect.New(),
	)
	if err := pg.Ping(); err != nil {
		return nil, errors.Wrapf(err, "cannot ping postgresql")
	}
	return pg, nil
}

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

	i.Operations = operations

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

func diffDefinitions(ctx context.Context, contractRepo core.ContractRepository, current map[abi.TLBType]abi.TLBFieldsDesc) (added, changed map[abi.TLBType]abi.TLBFieldsDesc, err error) {
	old, err := contractRepo.GetDefinitions(ctx)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get definitions")
	}

	added, changed = map[abi.TLBType]abi.TLBFieldsDesc{}, map[abi.TLBType]abi.TLBFieldsDesc{}
	for dt, d := range current {
		od, ok := old[dt]
		if !ok {
			added[dt] = d
		}
		if !reflect.DeepEqual(od, d) {
			changed[dt] = d
		}
	}

	return added, changed, nil
}

func diffSlices[V any](oldS, newS []V, getName func(v V) string) (added, changed, deleted []V) {
	oldM, newM := map[string]V{}, map[string]V{}
	for _, v := range oldS {
		oldM[getName(v)] = v
	}
	for _, v := range newS {
		newM[getName(v)] = v
	}

	for vn, v := range newM {
		ov, ok := oldM[vn]
		if !ok {
			added = append(added, v)
		}
		if !reflect.DeepEqual(ov, v) {
			changed = append(changed, v)
		}
	}
	for vn := range oldM {
		_, ok := newM[vn]
		if !ok {
			deleted = append(deleted, oldM[vn])
		}
	}

	return added, changed, deleted
}

func diffInterface(oldInterface, newInterface *core.ContractInterface) (interfaceChanged bool, added, changed, deleted []abi.GetMethodDesc) {
	interfaceChanged = !reflect.DeepEqual(newInterface.Addresses, oldInterface.Addresses) ||
		!reflect.DeepEqual(newInterface.Code, oldInterface.Code) ||
		!reflect.DeepEqual(newInterface.GetMethodHashes, oldInterface.GetMethodHashes)

	added, changed, deleted = diffSlices(oldInterface.GetMethodsDesc, newInterface.GetMethodsDesc, func(v abi.GetMethodDesc) string { return v.Name })

	return interfaceChanged, added, changed, deleted
}

func diffOperations(oldOperations, newOperations []*core.ContractOperation) (added, changed, deleted []*core.ContractOperation) {
	return diffSlices(oldOperations, newOperations, func(v *core.ContractOperation) string { return v.OperationName })
}

func getGetMethodNames(desc []abi.GetMethodDesc) (names []string) {
	for i := range desc {
		if len(desc[i].Arguments) > 0 {
			continue
		}
		names = append(names, desc[i].Name)
	}
	return
}

func rescanGetMethod(ctx context.Context, in abi.ContractName, repo core.RescanRepository, t core.RescanTaskType, getMethods []string) error {
	if len(getMethods) == 0 {
		return nil
	}

	err := repo.AddRescanTask(ctx, &core.RescanTask{
		Type:              t,
		ContractName:      in,
		ChangedGetMethods: getMethods,
	})
	if err != nil {
		return errors.Wrapf(err, "add rescan task for '%s' get-method", getMethods)
	}

	for _, gm := range getMethods {
		log.Info().
			Str("rescan_type", string(t)).
			Str("interface_name", string(in)).
			Str("get_method", gm).
			Msg("added get-method rescan task")
	}

	return nil
}

func rescanOperation(ctx context.Context, repo core.RescanRepository, t core.RescanTaskType, op *core.ContractOperation) error {
	err := repo.AddRescanTask(ctx, &core.RescanTask{
		Type:         t,
		ContractName: op.ContractName,
		MessageType:  op.MessageType,
		Outgoing:     op.Outgoing,
		OperationID:  op.OperationID,
	})
	if err != nil {
		return errors.Wrapf(err, "add rescan task for '%s' operation", op.OperationName)
	}

	log.Info().
		Str("rescan_type", string(t)).
		Str("interface_name", string(op.ContractName)).
		Str("operation_name", op.OperationName).
		Msg("added operation rescan task")

	return nil
}

var Command = &cli.Command{
	Name:  "contract",
	Usage: "Manages contract interfaces in the database",

	Subcommands: cli.Commands{
		{
			Name:  "addInterfaces",
			Usage: "Adds contract interface",

			ArgsUsage: "[file1.json] [file2.json]",

			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "stdin",
					Usage:   "read from stdin instead of files",
					Aliases: []string{"i"},
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

				pg, err := dbConnect()
				if err != nil {
					return err
				}

				contractRepo := contract.NewRepository(pg)
				rescanRepo := rescan.NewRepository(pg)

				addedDef, changedDef, err := diffDefinitions(ctx.Context, contractRepo, definitions)
				if err != nil {
					return err
				}
				for dn, d := range changedDef {
					if err := contractRepo.UpdateDefinition(ctx.Context, dn, d); err != nil {
						return errors.Wrapf(err, "cannot update contract definition '%s'", dn)
					}
				}
				for dn, d := range addedDef {
					if err := contractRepo.AddDefinition(ctx.Context, dn, d); err != nil {
						return errors.Wrapf(err, "cannot insert contract definition '%s'", dn)
					}
				}

				for _, i := range interfaces {
					if err := contractRepo.AddInterface(ctx.Context, i); err != nil {
						log.Error().Err(err).Str("interface_name", string(i.Name)).Msg("cannot insert contract interface")
						continue
					}
					err := rescanRepo.AddRescanTask(ctx.Context, &core.RescanTask{
						Type:         core.AddInterface,
						ContractName: i.Name,
					})
					if err != nil {
						log.Error().Err(err).Str("interface_name", string(i.Name)).Msg("cannot add interface rescan task")
					}
				}

				for _, op := range operations {
					if err := contractRepo.AddOperation(ctx.Context, op); err != nil {
						log.Error().Err(err).
							Str("interface_name", string(op.ContractName)).
							Str("operation_name", op.OperationName).
							Msg("cannot insert contract operation")
						continue
					}
					err := rescanRepo.AddRescanTask(ctx.Context, &core.RescanTask{
						Type:         core.UpdOperation,
						ContractName: op.ContractName,
						MessageType:  op.MessageType,
						Outgoing:     op.Outgoing,
						OperationID:  op.OperationID,
					})
					if err != nil {
						log.Error().Err(err).
							Str("interface_name", string(op.ContractName)).
							Str("op_name", op.OperationName).
							Msg("cannot add operation rescan task")
					}
				}

				return nil
			},
		},
		{
			Name:  "updateInterface",
			Usage: "Updates contract interface in the database and adds rescan tasks for the difference between old and new interfaces",

			ArgsUsage: "[file1.json] [file2.json]",

			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "stdin",
					Usage:   "read from stdin instead of files",
					Aliases: []string{"i"},
				},
				&cli.StringFlag{
					Name:     "contract-name",
					Usage:    "contract interface for update",
					Aliases:  []string{"c"},
					Required: true,
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

				definitions, interfaces, _, err := parseInterfacesDesc(interfacesDesc)
				if err != nil {
					return err
				}

				contractName := abi.ContractName(ctx.String("contract-name"))
				if contractName == "" {
					return errors.Wrap(core.ErrInvalidArg, "contract interface name is not set")
				}

				var newInterface *core.ContractInterface
				for _, i := range interfaces {
					if i.Name == contractName {
						newInterface = i
					}
				}
				if newInterface == nil {
					return errors.Wrapf(core.ErrInvalidArg, "contract interface '%s' is found in abi description", contractName)
				}

				pg, err := dbConnect()
				if err != nil {
					return err
				}

				contractRepo := contract.NewRepository(pg)
				rescanRepo := rescan.NewRepository(pg)

				oldInterface, err := contractRepo.GetInterface(ctx.Context, contractName)
				if err != nil {
					return errors.Wrapf(err, "get '%s' interface", newInterface.Name)
				}

				addedDef, changedDef, err := diffDefinitions(ctx.Context, contractRepo, definitions)
				if err != nil {
					return err
				}
				for dn, d := range changedDef {
					if err := contractRepo.UpdateDefinition(ctx.Context, dn, d); err != nil {
						return errors.Wrapf(err, "cannot update contract definition '%s'", dn)
					}
				}
				for dn, d := range addedDef {
					if err := contractRepo.AddDefinition(ctx.Context, dn, d); err != nil {
						return errors.Wrapf(err, "cannot insert contract definition '%s'", dn)
					}
				}

				iChanged, addedGm, changedGm, deletedGm := diffInterface(oldInterface, newInterface)
				if iChanged {
					if err := contractRepo.UpdateInterface(ctx.Context, newInterface); err != nil {
						return errors.Wrapf(err, "cannot update contract interface '%s'", newInterface.Name)
					}
				}

				addedOp, changedOp, deletedOp := diffOperations(oldInterface.Operations, newInterface.Operations)
				for _, op := range deletedOp {
					if err := contractRepo.DeleteOperation(ctx.Context, op.OperationName); err != nil {
						return errors.Wrapf(err, "cannot delete contract operation '%s'", op.OperationName)
					}
				}
				for _, op := range changedOp {
					if err := contractRepo.UpdateOperation(ctx.Context, op); err != nil {
						return errors.Wrapf(err, "cannot update contract operation '%s'", op.OperationName)
					}
				}
				for _, op := range addedOp {
					if err := contractRepo.AddOperation(ctx.Context, op); err != nil {
						return errors.Wrapf(err, "cannot insert contract operation '%s'", op.OperationName)
					}
				}

				if err := rescanGetMethod(ctx.Context, contractName, rescanRepo, core.AddGetMethod, getGetMethodNames(addedGm)); err != nil {
					return err
				}
				if err := rescanGetMethod(ctx.Context, contractName, rescanRepo, core.UpdGetMethod, getGetMethodNames(changedGm)); err != nil {
					return err
				}
				if err := rescanGetMethod(ctx.Context, contractName, rescanRepo, core.DelGetMethod, getGetMethodNames(deletedGm)); err != nil {
					return err
				}

				for _, op := range deletedOp {
					if err := rescanOperation(ctx.Context, rescanRepo, core.DelOperation, op); err != nil {
						return err
					}
				}
				for _, op := range append(addedOp, changedOp...) {
					if err := rescanOperation(ctx.Context, rescanRepo, core.UpdOperation, op); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			Name:  "deleteInterface",
			Usage: "Deletes contract interface from the database and removes associated parsed data",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "contract-name",
					Usage:    "contract interface for deletion",
					Aliases:  []string{"c"},
					Required: true,
				},
			},

			Action: func(ctx *cli.Context) (err error) {
				contractName := abi.ContractName(ctx.String("contract-name"))
				if contractName == "" {
					return errors.Wrap(core.ErrInvalidArg, "contract interface name is not set")
				}

				pg, err := dbConnect()
				if err != nil {
					return err
				}

				contractRepo := contract.NewRepository(pg)
				rescanRepo := rescan.NewRepository(pg)

				oldInterface, err := contractRepo.GetInterface(ctx.Context, contractName)
				if err != nil {
					return errors.Wrapf(err, "get '%s' interface", contractName)
				}

				if err := contractRepo.DeleteInterface(ctx.Context, contractName); err != nil {
					return errors.Wrapf(err, "cannot delete '%s' interface", contractName)
				}

				for _, op := range oldInterface.Operations {
					if err := rescanOperation(ctx.Context, rescanRepo, core.DelOperation, op); err != nil {
						return err
					}
				}

				err = rescanRepo.AddRescanTask(ctx.Context, &core.RescanTask{
					Type:         core.DelInterface,
					ContractName: contractName,
				})
				if err != nil {
					log.Error().Err(err).Str("interface_name", string(contractName)).Msg("cannot add interface rescan task")
				}

				return nil
			},
		},
	},
}
