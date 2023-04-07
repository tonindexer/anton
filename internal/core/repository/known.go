package repository

import (
	"context"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
)

func insertKnownInterfaces(ctx context.Context, repo core.ContractRepository) error {
	for n, get := range abi.KnownContractMethods {
		row := core.ContractInterface{
			Name:       n,
			GetMethods: get,
		}
		for _, g := range row.GetMethods {
			row.GetMethodHashes = append(row.GetMethodHashes, abi.MethodNameHash(g))
		}

		if err := repo.AddInterface(ctx, &row); err != nil {
			return errors.Wrapf(err, "%s [%v]", n, get)
		}
	}

	for v, code := range abi.WalletCode {
		row := core.ContractInterface{
			Name:     v.Name(),
			Code:     code.ToBOC(),
			CodeHash: code.Hash(),
		}
		if err := repo.AddInterface(ctx, &row); err != nil {
			return errors.Wrapf(err, "wallet code %s", row.Name)
		}
	}

	return nil
}

func insertKnownOperations(ctx context.Context, repo core.ContractRepository) error {
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
				if err := repo.AddOperation(ctx, &row); err != nil {
					return errors.Wrapf(err, "%s/%s", row.ContractName, row.Name)
				}
			}
		}
	}

	return nil
}

func insertKnownAddresses(ctx context.Context, repo core.ContractRepository) error {
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
		if err := repo.AddInterface(ctx, iface); err != nil {
			return errors.Wrapf(err, "%s [%v]", n, iface.Addresses)
		}
	}

	return nil
}

func InsertKnownInterfaces(ctx context.Context, repo core.ContractRepository) error {
	if err := insertKnownInterfaces(ctx, repo); err != nil {
		return err
	}

	if err := insertKnownOperations(ctx, repo); err != nil {
		return err
	}

	if err := insertKnownAddresses(ctx, repo); err != nil {
		return err
	}

	return nil
}
