package repository

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

func insertKnownInterfaces(ctx context.Context, db *ch.DB) error {
	var rows []*core.ContractInterface

	for n, get := range abi.KnownContractMethods {
		rows = append(rows, &core.ContractInterface{
			Name:       n,
			GetMethods: get,
		})
	}

	_, err := db.NewInsert().Model(&rows).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func insertKnownOperations(ctx context.Context, db *ch.DB) error {
	var rows []*core.ContractOperation

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

				rows = append(rows, &core.ContractOperation{
					Name:         strcase.ToSnake(reflect.TypeOf(msg).Name()),
					ContractName: n,
					Outgoing:     out,
					OperationID:  opID,
					Schema:       schema,
				})
			}
		}
	}

	_, err := db.NewInsert().Model(&rows).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func insertKnownAddresses(ctx context.Context, db *ch.DB) error {
	res, err := http.Get("https://raw.githubusercontent.com/menschee/tonscanplus/main/data.json")
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var addrMap = make(map[string]abi.ContractName)
	if err := json.Unmarshal(body, &addrMap); err != nil {
		return errors.Wrap(err, "tonscanplus data unmarshal")
	}

	var contracts []*core.ContractInterface
	for addr, name := range addrMap {
		contracts = append(contracts, &core.ContractInterface{
			Name:    name,
			Address: addr,
		})
	}

	_, err = db.NewInsert().Model(&contracts).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func InsertKnownInterfaces(ctx context.Context, db *ch.DB) error {
	if err := insertKnownInterfaces(ctx, db); err != nil {
		return err
	}

	if err := insertKnownAddresses(ctx, db); err != nil {
		return err
	}

	if err := insertKnownOperations(ctx, db); err != nil {
		return err
	}

	return nil
}
