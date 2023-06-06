package aggregate

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
)

type Statistics struct {
	FirstBlock int `json:"first_masterchain_block"`
	LastBlock  int `json:"last_masterchain_block"`
	BlockCount int `json:"masterchain_block_count"`

	AddressCount       int `json:"address_count"`
	ParsedAddressCount int `json:"parsed_address_count"`

	AccountCount       int `json:"account_count"`
	ParsedAccountCount int `json:"parsed_account_count"`

	TransactionCount int `json:"transaction_count"`

	MessageCount       int `json:"message_count"`
	ParsedMessageCount int `json:"parsed_message_count"`

	ContractInterfaceCount int `json:"contract_interface_count"`
	ContractOperationCount int `json:"contract_operation_count"`

	AccountStatusCount []struct {
		Status core.AccountStatus `json:"status"`
		Count  int                `json:"count"`
	} `json:"account_count_by_status"`

	AccountTypesCount []struct {
		Interfaces []abi.ContractName `json:"interfaces"`
		Count      int                `json:"count"`
	} `json:"account_count_by_interfaces"`

	MessageTypesCount []struct {
		Operation string `json:"operation"`
		Count     int    `json:"count"`
	} `json:"message_count_by_operation"`
}

func GetStatistics(ctx context.Context, ck *ch.DB, pg *bun.DB) (*Statistics, error) {
	var ret Statistics

	err := ck.NewSelect().Model((*core.Block)(nil)).
		ColumnExpr("min(seq_no) as first_masterchain_block").
		ColumnExpr("max(seq_no) as last_masterchain_block").
		Where("workchain = -1").
		Scan(ctx, &ret.FirstBlock, &ret.LastBlock)
	if err != nil {
		return nil, errors.Wrap(err, "first and last masterchain blocks")
	}
	ret.BlockCount, err = ck.NewSelect().Model((*core.Block)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "block count")
	}

	ret.AccountCount, err = ck.NewSelect().Model((*core.AccountState)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "account state count")
	}
	ret.ParsedAccountCount, err = ck.NewSelect().Model((*core.AccountState)(nil)).Where("length(types) > 0").Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "account data count")
	}

	err = ck.NewSelect().
		ColumnExpr("last_status as status").
		ColumnExpr("count(addr) as count").
		TableExpr("(?) as q",
			ck.NewSelect().
				Model((*core.AccountState)(nil)).
				ColumnExpr("argMax(address, last_tx_lt) as addr").
				ColumnExpr("argMax(status, last_tx_lt) as last_status").
				Group("address")).
		Group("last_status").
		Order("count DESC").
		Scan(ctx, &ret.AccountStatusCount)
	if err != nil {
		return nil, errors.Wrap(err, "account status count")
	}
	for _, row := range ret.AccountStatusCount {
		ret.AddressCount += row.Count
	}

	err = ck.NewSelect().
		ColumnExpr("last_types as interfaces").
		ColumnExpr("count(addr) as count").
		TableExpr("(?) as q",
			ck.NewSelect().
				Model((*core.AccountState)(nil)).
				ColumnExpr("argMax(address, last_tx_lt) as addr").
				ColumnExpr("argMax(types, last_tx_lt) as last_types").
				Group("address")).
		Group("last_types").
		Order("count DESC").
		Scan(ctx, &ret.AccountTypesCount)
	if err != nil {
		return nil, errors.Wrap(err, "account interfaces count")
	}
	for _, row := range ret.AccountTypesCount {
		if len(row.Interfaces) == 0 {
			continue
		}
		ret.ParsedAddressCount += row.Count
	}

	ret.TransactionCount, err = ck.NewSelect().Model((*core.Transaction)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "transaction count")
	}
	ret.MessageCount, err = ck.NewSelect().Model((*core.Message)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "message count")
	}
	err = ck.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("operation_name as operation").
		ColumnExpr("count() as count").
		Where("length(operation_name) > 0").
		Group("operation_name").
		Order("count DESC").
		Scan(ctx, &ret.MessageTypesCount)
	if err != nil {
		return nil, errors.Wrap(err, "account interfaces count")
	}
	for _, row := range ret.MessageTypesCount {
		ret.ParsedMessageCount += row.Count
	}

	ret.ContractInterfaceCount, err = pg.NewSelect().Model((*core.ContractInterface)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "contract interface count")
	}
	ret.ContractOperationCount, err = pg.NewSelect().Model((*core.ContractOperation)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "contract operation count")
	}

	return &ret, nil
}
