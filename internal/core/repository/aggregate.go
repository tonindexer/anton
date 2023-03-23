package repository

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

type Statistics struct {
	FirstBlock int `json:"first_masterchain_block"`
	LastBlock  int `json:"last_masterchain_block"`

	BlockCount int `json:"masterchain_block_count"`

	AddressCount       int `json:"address_count"`
	ParsedAddressCount int `json:"parsed_address_count"`
	AccountCount       int `json:"account_count"`
	AccountDataCount   int `json:"account_data_count"`
	AccountStatusCount []struct {
		Status core.AccountStatus `json:"status"`
		Count  int                `json:"count"`
	} `json:"account_count_by_status"`
	AccountTypesCount []struct {
		Interfaces []abi.ContractName `json:"interfaces"`
		Count      int                `json:"count"`
	} `json:"account_count_by_interfaces"`

	TransactionCount   int `json:"transaction_count"`
	MessageCount       int `json:"message_count"`
	ParsedMessageCount int `json:"parsed_message_count"`
	MessageTypesCount  []struct {
		Operation string `json:"operation"`
		Count     int    `json:"count"`
	} `json:"message_count_by_operation"`

	ContractInterfaceCount int `json:"contract_interface_count"`
	ContractOperationCount int `json:"contract_operation_count"`
}

func GetStatistics(ctx context.Context, db *DB) (*Statistics, error) {
	var ret Statistics

	err := db.CH.NewSelect().Model((*core.Block)(nil)).
		ColumnExpr("min(seq_no) as first_masterchain_block").
		ColumnExpr("max(seq_no) as last_masterchain_block").
		Where("workchain = -1").
		Scan(ctx, &ret.FirstBlock, &ret.LastBlock)
	if err != nil {
		return nil, errors.Wrap(err, "first and last masterchain blocks")
	}
	ret.BlockCount, err = db.CH.NewSelect().Model((*core.Block)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "block count")
	}

	ret.AccountCount, err = db.CH.NewSelect().Model((*core.AccountState)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "account state count")
	}
	err = db.CH.NewSelect().
		ColumnExpr("last_status as status").
		ColumnExpr("count(addr) as count").
		TableExpr("(?) as q",
			db.CH.NewSelect().
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

	ret.AccountDataCount, err = db.CH.NewSelect().Model((*core.AccountData)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "account data count")
	}
	err = db.CH.NewSelect().
		ColumnExpr("last_types as interfaces").
		ColumnExpr("count(addr) as count").
		TableExpr("(?) as q",
			db.CH.NewSelect().
				Model((*core.AccountData)(nil)).
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
		ret.ParsedAddressCount += row.Count
	}

	ret.TransactionCount, err = db.CH.NewSelect().Model((*core.Transaction)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "transaction count")
	}
	ret.MessageCount, err = db.CH.NewSelect().Model((*core.Message)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "message count")
	}
	err = db.CH.NewSelect().Model((*core.MessagePayload)(nil)).
		ColumnExpr("operation_name as operation").
		ColumnExpr("count() as count").
		Group("operation_name").
		Order("count DESC").
		Scan(ctx, &ret.MessageTypesCount)
	if err != nil {
		return nil, errors.Wrap(err, "account interfaces count")
	}
	for _, row := range ret.MessageTypesCount {
		ret.ParsedMessageCount += row.Count
	}

	ret.ContractInterfaceCount, err = db.PG.NewSelect().Model((*core.ContractInterface)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "contract interface count")
	}
	ret.ContractOperationCount, err = db.PG.NewSelect().Model((*core.ContractOperation)(nil)).Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "contract operation count")
	}

	return &ret, nil
}
