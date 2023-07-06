package aggregate

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
)

type AddressStatusCount struct {
	Status core.AccountStatus `json:"status"`
	Count  int                `json:"count"`
}

type AddressTypesCount struct {
	Interfaces []abi.ContractName `json:"interfaces"`
	Count      int                `json:"count"`
}

type Statistics struct {
	FirstBlock       int `json:"first_masterchain_block"`
	LastBlock        int `json:"last_masterchain_block"`
	MasterBlockCount int `json:"masterchain_block_count"`

	AddressCount       int `json:"address_count"`
	ParsedAddressCount int `json:"parsed_address_count"`

	AccountCount       int `json:"account_count"`
	ParsedAccountCount int `json:"parsed_account_count"`

	TransactionCount int `json:"transaction_count"`

	MessageCount       int `json:"message_count"`
	ParsedMessageCount int `json:"parsed_message_count"`

	ContractInterfaceCount int `json:"contract_interface_count"`
	ContractOperationCount int `json:"contract_operation_count"`

	AddressStatusCount []AddressStatusCount `json:"address_count_by_status"`

	AddressTypesCount []AddressTypesCount `json:"address_count_by_interfaces"`

	MessageTypesCount []struct {
		Operation string `json:"operation"`
		Count     int    `json:"count"`
	} `json:"message_count_by_operation"`
}

type accountCount struct {
	Parsed       uint8
	AccountCount int
}

type addressCount struct {
	Status     core.AccountStatus
	Interfaces []abi.ContractName
	Count      int
}

func marshalInterfacesKey(interfaces []abi.ContractName) string {
	var str []string
	for _, i := range interfaces {
		str = append(str, string(i))
	}
	sort.Strings(str)
	return strings.Join(str, ",")
}

func unmarshalInterfacesKey(raw string) (ret []abi.ContractName) {
	str := strings.Split(raw, ",")
	for _, s := range str {
		if s == "" {
			continue
		}
		ret = append(ret, abi.ContractName(s))
	}
	return ret
}

func getBlockStatistics(ctx context.Context, ck *ch.DB, ret *Statistics) error {
	err := ck.NewSelect().Model((*core.Block)(nil)).
		ColumnExpr("count(seq_no) as masterchain_block_count").
		ColumnExpr("min(seq_no) as first_masterchain_block").
		ColumnExpr("max(seq_no) as last_masterchain_block").
		Where("workchain = -1").
		Scan(ctx, &ret.MasterBlockCount, &ret.FirstBlock, &ret.LastBlock)
	if err != nil {
		return errors.Wrap(err, "first and last masterchain blocks")
	}
	return nil
}

func getAccountStatistics(ctx context.Context, ck *ch.DB, ret *Statistics) error {
	var accounts []accountCount

	err := ck.NewSelect().Model((*core.AccountState)(nil)).
		ColumnExpr("length(types) > 0 as parsed").
		ColumnExpr("count(address) as account_count").
		Group("parsed").
		Scan(ctx, &accounts)
	if err != nil {
		return errors.Wrap(err, "account state count")
	}
	for _, r := range accounts {
		ret.AccountCount += r.AccountCount
		if r.Parsed > 0 {
			ret.ParsedAccountCount += r.AccountCount
		}
	}

	var addresses []addressCount
	err = ck.NewSelect().
		ColumnExpr("last_status as status").
		ColumnExpr("last_types as interfaces").
		ColumnExpr("count(addr) as count").
		TableExpr("(?) as q",
			ck.NewSelect().
				Model((*core.AccountState)(nil)).
				ColumnExpr("argMax(address, last_tx_lt) as addr").
				ColumnExpr("argMax(status, last_tx_lt) as last_status").
				ColumnExpr("argMax(types, last_tx_lt) as last_types").
				Group("address")).
		Group("last_status", "last_types").
		Scan(ctx, &addresses)
	if err != nil {
		return errors.Wrap(err, "address count")
	}

	accountStatuses := map[core.AccountStatus]int{}
	accountTypes := map[string]int{}
	for _, r := range addresses {
		accountStatuses[r.Status] += r.Count
		accountTypes[marshalInterfacesKey(r.Interfaces)] += r.Count
	}

	for s, c := range accountStatuses {
		ret.AddressStatusCount = append(ret.AddressStatusCount, AddressStatusCount{Status: s, Count: c})
	}
	for i, c := range accountTypes {
		r := AddressTypesCount{Interfaces: unmarshalInterfacesKey(i), Count: c}
		if len(r.Interfaces) == 0 {
			continue
		}
		ret.AddressTypesCount = append(ret.AddressTypesCount, r)
	}

	sort.Slice(ret.AddressStatusCount, func(i, j int) bool { return ret.AddressStatusCount[i].Count > ret.AddressStatusCount[j].Count })
	sort.Slice(ret.AddressTypesCount, func(i, j int) bool { return ret.AddressTypesCount[i].Count > ret.AddressTypesCount[j].Count })

	for _, row := range ret.AddressStatusCount {
		ret.AddressCount += row.Count
	}
	for _, row := range ret.AddressTypesCount {
		if len(row.Interfaces) == 0 {
			continue
		}
		ret.ParsedAddressCount += row.Count
	}

	return nil
}

func getTransactionStatistics(ctx context.Context, ck *ch.DB, ret *Statistics) error {
	err := ck.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("operation_name as operation").
		ColumnExpr("count() as count").
		Group("operation_name").
		Order("count DESC").
		Scan(ctx, &ret.MessageTypesCount)
	if err != nil {
		return errors.Wrap(err, "message types count")
	}

	unknownOp := -1
	for it, row := range ret.MessageTypesCount {
		ret.MessageCount += row.Count
		if row.Operation != "" {
			ret.ParsedMessageCount += row.Count
		} else {
			unknownOp = it
		}
	}
	if unknownOp == len(ret.MessageTypesCount)-1 {
		ret.MessageTypesCount = ret.MessageTypesCount[:unknownOp]
	} else if unknownOp != -1 {
		ret.MessageTypesCount = append(ret.MessageTypesCount[:unknownOp], ret.MessageTypesCount[unknownOp+1:]...)
	}

	ret.TransactionCount, err = ck.NewSelect().Model((*core.Transaction)(nil)).Count(ctx)
	if err != nil {
		return errors.Wrap(err, "transaction count")
	}

	return nil
}

func GetStatistics(ctx context.Context, ck *ch.DB, pg *bun.DB) (*Statistics, error) {
	var (
		ret Statistics
		err error
	)

	if err := getBlockStatistics(ctx, ck, &ret); err != nil {
		return nil, err
	}

	if err := getAccountStatistics(ctx, ck, &ret); err != nil {
		return nil, err
	}

	if err := getTransactionStatistics(ctx, ck, &ret); err != nil {
		return nil, err
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
