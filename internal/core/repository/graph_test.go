package repository_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

func randBytes(l int) []byte {
	token := make([]byte, l)
	rand.Read(token) // nolint
	return token
}

func randAddr() *addr.Address {
	a, err := new(addr.Address).FromString(fmt.Sprintf("0:%x", randBytes(32)))
	if err != nil {
		panic(err)
	}
	return a
}

func randUint() uint64 {
	return rand.Uint64() // nolint
}

func randTs() int64 {
	return time.Now().Unix()
}

func randLT() uint64 {
	return uint64(rand.Uint32()) // nolint
}

func intFromStr(str string) *bunbig.Int {
	i, err := new(bunbig.Int).FromString(str)
	if err != nil {
		panic(err)
	}
	return i
}

var (
	master = core.Block{
		Workchain: -1,
		Shard:     -9223372036854775808,
		SeqNo:     1234,

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		Transactions: nil,
	}

	shard = core.Block{
		Workchain: 0,
		Shard:     -9223372036854775808,
		SeqNo:     4321,

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		MasterID: &core.BlockID{
			Workchain: master.Workchain,
			Shard:     master.Shard,
			SeqNo:     master.SeqNo,
		},

		Transactions: nil,
	}
	shardPrev = core.Block{
		Workchain: 0,
		Shard:     -9223372036854775808,
		SeqNo:     4320,

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		MasterID: &core.BlockID{
			Workchain: master.Workchain,
			Shard:     master.Shard,
			SeqNo:     master.SeqNo,
		},

		Transactions: nil,
	}

	addrWallet = *randAddr()

	accWalletOlder = core.AccountState{
		Address:    addrWallet,
		IsActive:   true,
		Status:     core.Active,
		Balance:    bunbig.FromInt64(1e9),
		LastTxLT:   randLT(),
		LastTxHash: randBytes(32),
		// Types:      []abi.ContractName{"wallet"},
	}

	accWalletOld = core.AccountState{
		Address:    accWalletOlder.Address,
		IsActive:   true,
		Status:     core.Active,
		Balance:    bunbig.FromInt64(31e9),
		LastTxLT:   accWalletOlder.LastTxLT + 1e3,
		LastTxHash: randBytes(32),
		// Types:      []abi.ContractName{"wallet"},
	}

	accWallet = core.AccountState{
		Address: accWalletOld.Address,

		IsActive: true,
		Status:   core.Active,
		Balance:  bunbig.FromInt64(3e9),

		LastTxLT:   accWalletOld.LastTxLT + 1e3,
		LastTxHash: randBytes(32),

		StateHash: randBytes(32),
		Code:      randBytes(128),
		CodeHash:  randBytes(32),
		Data:      randBytes(128), // parse it
		DataHash:  randBytes(32),

		// Types: []abi.ContractName{"wallet"},
	}

	addrItem = *randAddr()

	accItem = core.AccountState{
		Address: addrItem,

		IsActive: true,
		Status:   core.Active,
		Balance:  bunbig.FromInt64(54e9),

		LastTxLT:   accWallet.LastTxLT + 10,
		LastTxHash: randBytes(32),

		StateHash: randBytes(32),
		Code:      randBytes(128),
		CodeHash:  randBytes(32),
		Data:      randBytes(128), // parse it
		DataHash:  randBytes(32),

		// Types: []abi.ContractName{"item"},
		GetMethodHashes: []uint32{abi.MethodNameHash("get_item_data")},
	}

	addrNoState = randAddr()

	accNoState = core.AccountState{
		Address:    *addrNoState,
		IsActive:   false,
		Status:     core.NonExist,
		Balance:    bunbig.FromInt64(13),
		LastTxLT:   randLT(),
		LastTxHash: randBytes(32),
	}

	accDataWallet = core.AccountData{
		Address:    accWallet.Address,
		LastTxLT:   accWallet.LastTxLT,
		LastTxHash: accWallet.LastTxHash,
		Balance:    accWallet.Balance,
		Types:      []abi.ContractName{"wallet"},
	}

	ifaceItem = core.ContractInterface{
		Name:            accDataItem.Types[0],
		Addresses:       nil,
		Code:            nil,
		GetMethods:      []string{"get_item_data"},
		GetMethodHashes: []uint32{abi.MethodNameHash("get_item_data")},
	}

	idx, _ = new(bunbig.Int).FromString(fmt.Sprintf("%d", 43)) // TODO: bunbig.Int.FromUint64

	accDataItemJetBalance, _ = new(bunbig.Int).FromString(fmt.Sprintf("%d", randUint())) // TODO: bunbig.Int.FromUint64

	accDataItem = core.AccountData{
		Address:      accItem.Address,
		LastTxLT:     accItem.LastTxLT,
		LastTxHash:   accItem.LastTxHash,
		Balance:      accItem.Balance,
		Types:        []abi.ContractName{"item"},
		OwnerAddress: randAddr(),
		NFTCollectionData: core.NFTCollectionData{
			NextItemIndex: idx,
		},
		NFTRoyaltyData: core.NFTRoyaltyData{
			RoyaltyAddress: randAddr(),
		},
		NFTContentData: core.NFTContentData{
			ContentURI: "git://asdf.t",
		},
		NFTItemData: core.NFTItemData{
			ItemIndex:         bunbig.FromInt64(42),
			CollectionAddress: randAddr(),
		},
		FTWalletData: core.FTWalletData{JettonBalance: accDataItemJetBalance},
	}

	msgExtWallet = core.Message{
		Type:          core.ExternalIn,
		Hash:          randBytes(32),
		DstAddress:    accWallet.Address,
		Body:          randBytes(128),
		BodyHash:      randBytes(32),
		StateInitCode: nil,
		StateInitData: nil,
		CreatedAt:     time.Unix(randTs(), 0).UTC(),
		CreatedLT:     accWallet.LastTxLT,
	}

	txOutWallet = core.Transaction{
		Address: accWallet.Address,
		Hash:    accWallet.LastTxHash,

		BlockWorkchain: shard.Workchain,
		BlockShard:     shard.Shard,
		BlockSeqNo:     shard.SeqNo,

		PrevTxHash: randBytes(32),
		PrevTxLT:   randLT(),

		InMsgHash: msgExtWallet.Hash,
		InAmount:  bunbig.FromInt64(0),
		OutAmount: intFromStr("999999999800000000000"),

		TotalFees:   bunbig.FromInt64(1e5),
		StateUpdate: nil,
		Description: nil,
		OrigStatus:  core.Active,
		EndStatus:   core.Active,

		CreatedAt: msgExtWallet.CreatedAt,
		CreatedLT: accWallet.LastTxLT,
	}

	msgOutWallet = core.Message{
		Type: core.Internal,
		Hash: randBytes(32),

		SourceTxHash: txOutWallet.Hash,
		SourceTxLT:   txOutWallet.CreatedLT,

		SrcAddress: accWallet.Address,
		DstAddress: accItem.Address,

		Amount: txOutWallet.OutAmount,

		IHRDisabled: false,
		IHRFee:      bunbig.FromInt64(0),
		FwdFee:      bunbig.FromInt64(0),

		Body:        randBytes(32),
		BodyHash:    randBytes(32),
		OperationID: 0xffeeee,

		CreatedAt: msgExtWallet.CreatedAt,
		CreatedLT: accWallet.LastTxLT + 1,
	}

	txInItem = core.Transaction{
		Address: accItem.Address,
		Hash:    accItem.LastTxHash,

		BlockWorkchain: shard.Workchain,
		BlockShard:     shard.Shard,
		BlockSeqNo:     shard.SeqNo,

		PrevTxHash: randBytes(32),
		PrevTxLT:   randLT(),

		InMsgHash: msgOutWallet.Hash,
		InAmount:  msgOutWallet.Amount,
		OutAmount: bunbig.FromInt64(0),

		TotalFees: bunbig.FromInt64(1e3),

		StateUpdate: randBytes(32),
		Description: randBytes(32),

		OrigStatus: core.Active,
		EndStatus:  core.Active,

		CreatedAt: msgOutWallet.CreatedAt,
		CreatedLT: accItem.LastTxLT, // msgInItem.CreatedLT + 1,
	}

	opItemTransfer = core.ContractOperation{
		Name:         msgInItemPayload.OperationName,
		ContractName: msgInItemPayload.DstContract,
		Outgoing:     false,
		OperationID:  msgInItemPayload.OperationID,
		Schema:       []byte(`{}`),
	}

	msgInItemPayload = core.MessagePayload{
		Type: core.Internal,
		Hash: msgOutWallet.Hash,

		SrcAddress:  msgOutWallet.SrcAddress,
		SrcContract: accDataWallet.Types[0],
		DstAddress:  msgOutWallet.DstAddress,
		DstContract: accDataItem.Types[0],

		BodyHash:      msgOutWallet.BodyHash,
		OperationID:   msgOutWallet.OperationID,
		OperationName: "item_transfer",
		DataJSON:      json.RawMessage(`{"new_owner": "kkkkkk", "collection_address": "aaaaaa"}`),

		CreatedAt: msgOutWallet.CreatedAt,
		CreatedLT: msgOutWallet.CreatedLT,
	}
)
