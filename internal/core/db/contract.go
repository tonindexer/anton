package db

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository/account"
)

func insertInterfacesNFT(ctx context.Context, db *ch.DB) error {
	nft := []*core.ContractInterface{
		{
			Name:       core.NFTCollection,
			GetMethods: []string{"get_collection_data", "get_nft_address_by_index", "get_nft_content"},
		}, {
			Name:       core.NFTRoyalty,
			GetMethods: []string{"royalty_params"},
		}, {
			Name:       core.NFTEditable,
			GetMethods: []string{"get_editor"},
		}, {
			Name:       core.NFTItem,
			GetMethods: []string{"get_nft_data"},
		}, {
			Name:       core.NFTItemSBT,
			GetMethods: []string{"get_nft_data", "get_nonce", "get_public_key", "get_authority_address"},
		},
	}

	_, err := db.NewInsert().Model(&nft).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func insertOperationsNFT(ctx context.Context, db *ch.DB) error {
	// TODO: here is only incoming message, parse outcoming too

	operations := []*core.ContractOperation{{
		Name:         "nft_collection_mint",
		ContractName: core.NFTCollection,
		OperationID:  1,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000001"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "Index", Type: reflect.TypeOf(big.NewInt(0)), Tag: `tlb:"## 64"`},
			{Name: "TonAmount", Type: reflect.TypeOf(tlb.Coins{}), Tag: `tlb:"."`},
			// Content
		},
	}, {
		Name:         "nft_collection_mint_batch",
		ContractName: core.NFTCollection,
		OperationID:  2,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000002"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
		},
	}, {
		Name:         "nft_collection_change_owner",
		ContractName: core.NFTCollection,
		OperationID:  3,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000003"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "NewOwner", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
		},
	}, {
		Name:         "nft_collection_change_content",
		ContractName: core.NFTCollection,
		OperationID:  4,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000004"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			// Content
		},
	}, {
		Name:         "nft_item_transfer",
		ContractName: core.NFTItem,
		OperationID:  0x5fcc3d14,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#5fcc3d14"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "NewOwner", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
			{Name: "ResponseDestination", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
			//	CustomPayload       *cell.Cell       `tlb:"maybe ^"`
			//	ForwardAmount       tlb.Coins        `tlb:"."`
			//	ForwardPayload      *cell.Cell       `tlb:"either . ^"`
		},
	}, {
		Name:         "nft_item_change_content",
		ContractName: core.NFTEditable,
		OperationID:  0x1a0b9d51,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#1a0b9d51"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			// Content
		},
	}}

	accRepo := account.NewRepository(db)
	if err := accRepo.InsertContractOperations(ctx, operations); err != nil {
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

	var addrMap = make(map[string]string)
	if err := json.Unmarshal(body, &addrMap); err != nil {
		return errors.Wrap(err, "tonscanplus data unmarshal")
	}

	var contracts []*core.ContractInterface
	for addr, name := range addrMap {
		contracts = append(contracts, &core.ContractInterface{
			Name:    core.ContractType(name),
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
	if err := insertInterfacesNFT(ctx, db); err != nil {
		return err
	}

	if err := insertKnownAddresses(ctx, db); err != nil {
		return err
	}

	if err := insertOperationsNFT(ctx, db); err != nil {
		return err
	}

	return nil
}
