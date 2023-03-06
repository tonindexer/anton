package repository

import (
	"context"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

func insertKnownInterfaces(ctx context.Context, db *bun.DB) error {
	for n, get := range abi.KnownContractMethods {
		row := core.ContractInterface{
			Name:       n,
			GetMethods: get,
		}
		for _, g := range row.GetMethods {
			row.GetMethodHashes = append(row.GetMethodHashes, abi.MethodNameHash(g))
		}
		_, err := db.NewInsert().Model(&row).Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "%s [%v]", n, get)
		}
	}

	for v, code := range abi.WalletCode {
		row := core.ContractInterface{
			Name:     v.Name(),
			Code:     code.ToBOC(),
			CodeHash: code.Hash(),
		}
		_, err := db.NewInsert().Model(&row).Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "wallet code %s", row.Name)
		}
	}

	return nil
}

func insertKnownOperations(ctx context.Context, db *bun.DB) error {
	for n, m := range abi.KnownContractOperations {
		for out, messages := range m {
			for _, msg := range messages {
				schema, err := abi.MarshalSchema(msg)
				if err != nil {
					return errors.Wrap(err, "marshal schema")
				}

				opID, err := abi.OperationID(msg)
				if err != nil {
					return errors.Wrap(err, "get operation id")
				}

				row := core.ContractOperation{
					Name:         strcase.ToSnake(reflect.TypeOf(msg).Elem().Name()),
					ContractName: n,
					Outgoing:     out,
					OperationID:  opID,
					Schema:       schema,
				}
				_, err = db.NewInsert().Model(&row).Exec(ctx)
				if err != nil {
					return errors.Wrapf(err, "%s/%s", row.ContractName, row.Name)
				}
			}
		}
	}

	return nil
}

func insertKnownAddresses(ctx context.Context, db *bun.DB) error {
	var addrMap = make(map[string]abi.ContractName)

	// res, err := http.Get("https://raw.githubusercontent.com/menschee/tonscanplus/main/data.json")
	// if err != nil {
	// 	return err
	// }
	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }
	// if err := json.Unmarshal(body, &addrMap); err != nil {
	// 	return errors.Wrap(err, "tonscanplus data unmarshal")
	// }

	for a, n := range abi.KnownAddresses {
		addrMap[a] = n
	}

	knownAddr := make(map[abi.ContractName]*core.ContractInterface)
	for a, n := range addrMap {
		if knownAddr[n] == nil {
			knownAddr[n] = new(core.ContractInterface)
			knownAddr[n].Name = n
		}
		knownAddr[n].Addresses = append(knownAddr[n].Addresses, addr.MustFromBase64(a))
	}
	for n, iface := range knownAddr {
		_, err := db.NewInsert().Model(iface).Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "%s [%v]", n, iface.Addresses)
		}
	}

	return nil
}

func InsertKnownInterfaces(ctx context.Context, db *bun.DB) error {
	if err := insertKnownInterfaces(ctx, db); err != nil {
		return err
	}

	if err := insertKnownOperations(ctx, db); err != nil {
		return err
	}

	if err := insertKnownAddresses(ctx, db); err != nil {
		return err
	}

	return nil
}
