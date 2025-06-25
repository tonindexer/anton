package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (r *Repository) filterAddressLabels(ctx context.Context, f *filter.LabelsReq) (ret []*core.AddressLabel, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if len(f.Categories) > 0 {
		q = q.Where("categories && ?", pgdialect.Array(f.Categories))
	}

	q = q.Order("name ASC")

	q = q.Offset(f.Offset)

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)

	return ret, err
}

func (r *Repository) countAddressLabels(ctx context.Context, f *filter.LabelsReq) (int, error) {
	q := r.ch.NewSelect().Model((*core.AddressLabel)(nil))

	if f.Name != "" {
		q = q.Where("positionCaseInsensitive(name, ?) > 0", f.Name)
	}
	if len(f.Categories) > 0 {
		q = q.Where("hasAny(categories, ?)", ch.Array(f.Categories))
	}

	return q.Count(ctx)
}

func (r *Repository) FilterLabels(ctx context.Context, f *filter.LabelsReq) (*filter.LabelsRes, error) {
	var (
		res = new(filter.LabelsRes)
		err error
	)

	res.Rows, err = r.filterAddressLabels(ctx, f)
	if err != nil {
		if strings.Contains(err.Error(), "invalid input value for enum label_category") {
			return nil, errors.Wrap(core.ErrInvalidArg, "invalid input value for enum label_category")
		}
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countAddressLabels(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}

func flattenStateIDs(ids []*core.AccountStateID) (ret [][]any) {
	for _, id := range ids {
		ret = append(ret, []any{&id.Address, id.LastTxLT})
	}
	return
}

func (r *Repository) filterAccountStates(ctx context.Context, f *filter.AccountsReq) (ret []*core.AccountState, err error) {
	var (
		q                   *bun.SelectQuery
		prefix, statesTable string
		latest              []*core.LatestAccountState
	)

	// choose table to filter states by
	// and optionally join account data
	if f.LatestState {
		q = r.pg.NewSelect().Model(&latest).
			Relation("AccountState", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.ExcludeColumn(f.ExcludeColumn...)
			}).
			Relation("AccountState.Label")
		statesTable = "latest_account_state."
		prefix = "account_state."
	} else {
		q = r.pg.NewSelect().Model(&ret).
			ExcludeColumn(f.ExcludeColumn...).
			Relation("Label")
		statesTable = "account_state."
	}

	if len(f.Addresses) > 0 {
		q = q.Where(statesTable+"address in (?)", bun.In(f.Addresses))
	}
	if len(f.StateIDs) > 0 {
		q = q.Where("("+statesTable+"address, "+statesTable+"last_tx_lt) IN (?)", bun.In(flattenStateIDs(f.StateIDs)))
	}

	if f.Workchain != nil {
		q = q.Where(prefix+"workchain = ?", *f.Workchain)
	}
	if f.Shard != nil {
		q = q.Where(prefix+"shard = ?", *f.Shard)
	}
	if f.BlockSeqNoLeq != nil {
		q = q.Where(prefix+"block_seq_no <= ?", *f.BlockSeqNoLeq)
	}
	if f.BlockSeqNoBeq != nil {
		q = q.Where(prefix+"block_seq_no >= ?", *f.BlockSeqNoBeq)
	}

	if len(f.ContractTypes) > 0 {
		q = q.Where(statesTable+"types && ?", pgdialect.Array(f.ContractTypes))
	}
	if f.OwnerAddress != nil {
		q = q.Where(statesTable+"owner_address = ?", f.OwnerAddress)
	}
	if f.MinterAddress != nil {
		q = q.Where(statesTable+"minter_address = ?", f.MinterAddress)
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where(statesTable+"last_tx_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where(statesTable+"last_tx_lt < ?", f.AfterTxLT)
		}
	}
	if f.Order != "" {
		orderBy := "last_tx_lt"
		if f.BlockSeqNoLeq != nil || f.BlockSeqNoBeq != nil {
			orderBy = "block_seq_no"
		}
		q = q.Order(statesTable + orderBy + " " + strings.ToUpper(f.Order))
	}

	// if total < 100000 && f.LatestState {
	// 	// firstly, select all latest states, then apply limit
	// 	// https://ottertune.com/blog/how-to-fix-slow-postgresql-queries
	// 	rawQuery := fmt.Sprintf("WITH q AS MATERIALIZED (?) SELECT * FROM q LIMIT %d", f.Limit)
	// 	err = r.pg.NewRaw(rawQuery, q).Scan(ctx, &ret)
	// } else {
	err = q.Limit(f.Limit).Scan(ctx)
	// }

	if f.LatestState {
		for _, a := range latest {
			ret = append(ret, a.AccountState)
		}
	}

	return ret, err
}

func (r *Repository) countAccountStatesFullScan(ctx context.Context, f *filter.AccountsReq) (count int, maxLt uint64, err error) {
	if f.LatestState && (len(f.ContractTypes) > 0 || f.MinterAddress != nil || f.OwnerAddress != nil) {
		return 0, 0, errors.New("clickhouse latest account states full scan is not supported for these filters")
	}

	var result struct {
		Count int
		MaxLT *uint64 `ch:"max_lt"`
	}

	var q *ch.SelectQuery
	if f.LatestState {
		// For latest account states, we need to count distinct addresses
		q = r.ch.NewSelect().
			Model((*core.AccountState)(nil)).
			ColumnExpr("count(distinct address) AS count")
	} else {
		// For historical account states, we count all records
		q = r.ch.NewSelect().
			Model((*core.AccountState)(nil)).
			ColumnExpr("count(*) AS count")
	}
	q = q.ColumnExpr("floor((SELECT max(last_tx_lt) AS v FROM account_states), -7) - 1e7 as max_lt") // unfiltered max
	q = q.Where("last_tx_lt <= max_lt")

	if len(f.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(f.Addresses))
	}
	if len(f.StateIDs) > 0 {
		return 0, 0, errors.Wrap(core.ErrNotImplemented, "do not count on filter by account state ids")
	}

	if f.Workchain != nil {
		q = q.Where("workchain = ?", *f.Workchain)
	}
	if f.Shard != nil {
		q = q.Where("shard = ?", *f.Shard)
	}
	if f.BlockSeqNoLeq != nil {
		q = q.Where("block_seq_no <= ?", *f.BlockSeqNoLeq)
	}
	if f.BlockSeqNoBeq != nil {
		q = q.Where("block_seq_no >= ?", *f.BlockSeqNoBeq)
	}

	if len(f.ContractTypes) > 0 {
		q = q.Where("hasAny(types, ?)", ch.Array(f.ContractTypes))
	}
	if f.MinterAddress != nil {
		q = q.Where("minter_address = ?", f.MinterAddress)
	}
	if f.OwnerAddress != nil {
		q = q.Where("owner_address = ?", f.OwnerAddress)
	}

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == nil {
		return 0, 0, core.ErrNotFound
	}

	return result.Count, *result.MaxLT, nil
}

func (r *Repository) countLatestAccountStatesFullScanFiltered(ctx context.Context, req *filter.AccountsReq) (count int, err error) {
	q := r.pg.NewSelect().Model((*core.LatestAccountState)(nil)).
		ColumnExpr("count(*) AS count")

	if len(req.Addresses) > 0 {
		q = q.Where("address in (?)", bun.In(req.Addresses))
	}

	if len(req.ContractTypes) > 0 {
		q = q.Where("types && ?", pgdialect.Array(req.ContractTypes))
	}
	if req.OwnerAddress != nil {
		q = q.Where("owner_address = ?", req.OwnerAddress)
	}
	if req.MinterAddress != nil {
		q = q.Where("minter_address = ?", req.MinterAddress)
	}

	if err := q.Scan(ctx, &count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) countAccountStatesPartialScan(ctx context.Context, req *filter.AccountsReq, startLt uint64) (partialCount, roundedCount int, roundedMaxLt uint64, err error) {
	var result struct {
		SinceStartCount int    `bun:"since_start_count"`
		RoundedCount    int    `bun:"until_rounded_count"`
		RoundedMaxLT    uint64 `bun:"rounded_max_lt_value"`
	}

	selectTable := func() *bun.SelectQuery {
		if req.LatestState {
			return r.pg.NewSelect().Model((*core.LatestAccountState)(nil))
		} else {
			return r.pg.NewSelect().Model((*core.AccountState)(nil))
		}
	}
	ltColumn := func() string {
		if req.LatestState {
			return "created_lt"
		} else {
			return "last_tx_lt"
		}
	}
	applyFilters := func(q *bun.SelectQuery) *bun.SelectQuery {
		if len(req.Addresses) > 0 {
			q = q.Where("address in (?)", bun.In(req.Addresses))
		}
		if !req.LatestState {
			if req.Workchain != nil {
				q = q.Where("workchain = ?", *req.Workchain)
			}
			if req.Shard != nil {
				q = q.Where("shard = ?", *req.Shard)
			}
			if req.BlockSeqNoLeq != nil {
				q = q.Where("block_seq_no <= ?", *req.BlockSeqNoLeq)
			}
			if req.BlockSeqNoBeq != nil {
				q = q.Where("block_seq_no >= ?", *req.BlockSeqNoBeq)
			}
		}
		if len(req.ContractTypes) > 0 {
			q = q.Where("types && ?", pgdialect.Array(req.ContractTypes))
		}
		if req.OwnerAddress != nil {
			q = q.Where("owner_address = ?", req.OwnerAddress)
		}
		if req.MinterAddress != nil {
			q = q.Where("minter_address = ?", req.MinterAddress)
		}
		return q
	}

	q := r.pg.NewSelect().
		With(
			"rounded_max_lt",
			selectTable().
				ColumnExpr(fmt.Sprintf("floor(max(%s) / 1e7) * 1e7 - 1e7 AS v", ltColumn())),
		).
		With(
			"until_rounded_count",
			applyFilters(selectTable()).
				Table("rounded_max_lt").
				ColumnExpr("count(*) as v").
				Where(ltColumn()+" > ?", startLt).
				Where(ltColumn()+" <= rounded_max_lt.v"),
		).
		With(
			"since_start_count",
			applyFilters(selectTable()).
				ColumnExpr("count(*) as v").
				Where(ltColumn()+" > ?", startLt),
		).
		Table("rounded_max_lt", "until_rounded_count", "since_start_count").
		ColumnExpr("since_start_count.v AS since_start_count").
		ColumnExpr("until_rounded_count.v AS until_rounded_count").
		ColumnExpr("rounded_max_lt.v as rounded_max_lt_value")

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, 0, err
	}

	return result.SinceStartCount, result.RoundedCount, result.RoundedMaxLT, nil
}

func (r *Repository) countAccountStates(ctx context.Context, req *filter.AccountsReq) (int, error) {
	if req.LatestState && (len(req.ContractTypes) > 0 || req.OwnerAddress != nil || req.MinterAddress != nil) {
		count, err := r.countLatestAccountStatesFullScanFiltered(ctx, req)
		return count, err
	}

	// choose the appropriate cache based on whether we're querying latest or historical states
	cache := r.statesFilterCountCache
	if req.LatestState {
		cache = r.latestStatesFilterCountCache
	}

	// try to get from cache
	count, maxLT, err := cache.Get(req.AccountsFilter)
	if errors.Is(err, core.ErrNotFound) {
		// full scan for initial count
		count, maxLT, err = r.countAccountStatesFullScan(ctx, req)
		if errors.Is(err, core.ErrNotFound) {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		if err := cache.Set(req.AccountsFilter, count, maxLT); err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	// get partial count since last cached value
	partialCount, roundedPartialCount, roundedMaxLT, err := r.countAccountStatesPartialScan(ctx, req, maxLT)
	if err != nil {
		return 0, err
	}
	if err := cache.Set(req.AccountsFilter, count+roundedPartialCount, roundedMaxLT); err != nil {
		return 0, err
	}

	return count + partialCount, nil
}

func (r *Repository) getCodeData(ctx context.Context, rows []*core.AccountState, excludeCode, excludeData bool) error { //nolint:gocognit,gocyclo // TODO: make one function working for both code and data
	defer core.Timer(time.Now(), "getCodeData(%d, %t, %t)", len(rows), excludeCode, excludeData)

	codeHashesSet, dataHashesSet := map[string]struct{}{}, map[string]struct{}{}
	for _, row := range rows {
		if !excludeCode && len(row.Code) == 0 && len(row.CodeHash) == 32 {
			codeHashesSet[string(row.CodeHash)] = struct{}{}
		}
		if !excludeData && len(row.Data) == 0 && len(row.DataHash) == 32 {
			dataHashesSet[string(row.DataHash)] = struct{}{}
		}
	}

	batchLen := 1000
	codeHashBatches, dataHashBatches := make([][][]byte, 1), make([][][]byte, 1)
	appendHash := func(hash []byte, batches [][][]byte) [][][]byte {
		b := batches[len(batches)-1]
		if len(b) >= batchLen {
			b = [][]byte{}
			batches = append(batches, b)
		}
		batches[len(batches)-1] = append(b, hash)
		return batches
	}
	for h := range codeHashesSet {
		codeHashBatches = appendHash([]byte(h), codeHashBatches)
	}
	for h := range dataHashesSet {
		dataHashBatches = appendHash([]byte(h), dataHashBatches)
	}

	codeRes, dataRes := map[string][]byte{}, map[string][]byte{}
	for _, b := range codeHashBatches {
		var codeArr []*core.AccountStateCode
		err := r.ch.NewSelect().Model(&codeArr).Where("code_hash IN ?", ch.In(b)).Scan(ctx)
		if err != nil {
			return errors.Wrapf(err, "get code")
		}
		for _, x := range codeArr {
			codeRes[string(x.CodeHash)] = x.Code
		}
	}
	for _, b := range dataHashBatches {
		var dataArr []*core.AccountStateData
		err := r.ch.NewSelect().Model(&dataArr).Where("data_hash IN ?", ch.In(b)).Scan(ctx)
		if err != nil {
			return errors.Wrapf(err, "get data")
		}
		for _, x := range dataArr {
			dataRes[string(x.DataHash)] = x.Data
		}
	}

	for _, row := range rows {
		var ok bool
		if !excludeCode && len(row.Code) == 0 && len(row.CodeHash) == 32 {
			if row.Code, ok = codeRes[string(row.CodeHash)]; !ok {
				return fmt.Errorf("cannot find code with %x hash", row.CodeHash)
			}
		}
		if !excludeData && len(row.Data) == 0 && len(row.DataHash) == 32 {
			if row.Data, ok = dataRes[string(row.DataHash)]; !ok {
				return fmt.Errorf("cannot find data with %x hash", row.DataHash)
			}
		}
	}

	return nil
}

func (r *Repository) FilterAccounts(ctx context.Context, f *filter.AccountsReq) (*filter.AccountsRes, error) {
	var (
		res = new(filter.AccountsRes)
		err error
	)

	if f.Limit == 0 {
		f.Limit = 3
	}

	if f.Count {
		res.Total, err = r.countAccountStates(ctx, f)
		if err != nil && !errors.Is(err, core.ErrNotImplemented) {
			return res, errors.Wrap(err, "count account states")
		}
		if res.Total == 0 && !errors.Is(err, core.ErrNotImplemented) {
			return res, nil
		}
	}

	res.Rows, err = r.filterAccountStates(ctx, f)
	if err != nil {
		return res, err
	}

	var excludeCode, excludeData bool
	for _, c := range f.ExcludeColumn {
		cl := strings.ToLower(c)
		if cl == "code" {
			excludeCode = true
		}
		if cl == "data" {
			excludeData = true
		}
	}
	if f.WithCodeData && (!excludeCode || !excludeData) {
		if err := r.getCodeData(ctx, res.Rows, excludeCode, excludeData); err != nil {
			return res, err
		}
	}

	return res, nil
}
