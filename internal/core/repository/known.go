package repository

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
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
		}, {
			Name:       core.NFTSale,
			GetMethods: []string{"get_sale_data"},
		},
	}

	_, err := db.NewInsert().Model(&nft).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func insertOperationsNFT(ctx context.Context, _ch *ch.DB, _ *bun.DB) error {
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
		Name:         "nft_item_ownership_assigned",
		ContractName: core.NFTItem,
		Outgoing:     true,
		OperationID:  0x05138d91,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#05138d91"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "OwnerAddress", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
		},
	}, {
		Name:         "nft_item_excesses",
		ContractName: core.NFTItem,
		Outgoing:     true,
		OperationID:  0xd53276db,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#d53276db"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
		},
	}, {
		Name:         "nft_item_get_static_data",
		ContractName: core.NFTItem,
		OperationID:  0x2fcb26a2,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#2fcb26a2"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
		},
	}, {
		Name:         "nft_item_report_static_data",
		ContractName: core.NFTItem,
		Outgoing:     true,
		OperationID:  0x8b771735,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#8b771735"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "Index", Type: reflect.TypeOf(big.NewInt(0)), Tag: `tlb:"## 256"`},
			{Name: "OwnerAddress", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
		},
	}, {
		Name:         "nft_collection_get_royalty_params",
		ContractName: core.NFTCollection,
		OperationID:  0x693d3950,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#693d3950"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
		},
	}, {
		Name:         "nft_collection_report_royalty_params",
		ContractName: core.NFTCollection,
		Outgoing:     true,
		OperationID:  0xa8cb00ad,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#a8cb00ad"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
		},
	}, {
		Name:         "nft_item_edit_content",
		ContractName: core.NFTEditable,
		OperationID:  0x1a0b9d51,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#1a0b9d51"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			// Content
		},
	}, {
		Name:         "nft_item_transfer_editorship",
		ContractName: core.NFTEditable,
		OperationID:  0x1c04412a,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#1c04412a"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "NewEditor", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
			{Name: "ResponseDestination", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
		},
	}, {
		Name:         "nft_item_editorship_assigned",
		ContractName: core.NFTEditable,
		Outgoing:     true,
		OperationID:  0x511a4463,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#511a4463"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "EditorAddress", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
		},
	}, {
		Name:         "nft_sale_accept_coins",
		ContractName: core.NFTSale,
		OperationID:  1,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000001"`},
		},
	}, {
		Name:         "nft_sale_buy",
		ContractName: core.NFTSale,
		OperationID:  2,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000002"`},
		},
	}, {
		Name:         "nft_sale_cancel_sale",
		ContractName: core.NFTSale,
		OperationID:  3,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#00000003"`},
		},
	}}

	if err := contract.NewRepository(_ch).InsertOperations(ctx, operations); err != nil {
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

	if err := insertOperationsNFT(ctx, db, nil); err != nil {
		return err
	}

	return nil
}
