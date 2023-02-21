package repository_test

import (
	"fmt"
	"math/rand"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

func randBytes(l int) []byte {
	token := make([]byte, l)
	rand.Read(token) // nolint
	return token
}

func randAddr() string {
	return fmt.Sprintf("0:%x", randBytes(32))
}

func randUint() uint64 {
	return rand.Uint64() // nolint
}

func randTs() uint64 {
	return randUint()
}

func randLT() uint64 {
	return randUint()
}

var (
	master = core.Block{
		BlockID: core.BlockID{
			Workchain: -1,
			Shard:     2222,
			SeqNo:     1234,
		},

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		Transactions: nil,
	}

	shard = core.Block{
		BlockID: core.BlockID{
			Workchain: 0,
			Shard:     8888,
			SeqNo:     4321,
		},

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		MasterID: master.BlockID,

		Transactions: nil,
	}
	shardPrev = core.Block{
		BlockID: core.BlockID{
			Workchain: 0,
			Shard:     8888,
			SeqNo:     4320,
		},

		// FileHash: string(randBytes(32)),
		FileHash: randBytes(32),
		RootHash: randBytes(32),

		MasterID: master.BlockID,

		Transactions: nil,
	}

	accWalletOlder = core.AccountState{
		Latest:     true,
		Address:    randAddr(),
		IsActive:   true,
		Status:     core.Active,
		Balance:    1e9,
		LastTxLT:   randLT(),
		LastTxHash: randBytes(32),
		Types:      []string{"wallet"},
	}

	accWalletOld = core.AccountState{
		Latest:     true,
		Address:    accWalletOlder.Address,
		IsActive:   true,
		Status:     core.Active,
		Balance:    1e9,
		LastTxLT:   accWalletOlder.LastTxLT + 1e3,
		LastTxHash: randBytes(32),
		Types:      []string{"wallet"},
	}

	accWallet = core.AccountState{
		Latest:  true,
		Address: accWalletOld.Address,

		IsActive: true,
		Status:   core.Active,
		Balance:  1e9,

		LastTxLT:   accWalletOld.LastTxLT + 1e3,
		LastTxHash: randBytes(32),

		StateHash: randBytes(32),
		Code:      randBytes(128),
		CodeHash:  randBytes(32),
		Data:      randBytes(128), // parse it
		DataHash:  randBytes(32),

		Types: []string{"wallet"},
	}

	accItem = core.AccountState{
		Latest:  true,
		Address: randAddr(),

		IsActive: true,
		Status:   core.Active,
		Balance:  1e9,

		LastTxLT:   accWallet.LastTxLT + 10,
		LastTxHash: randBytes(32),

		StateHash: randBytes(32),
		Code:      randBytes(128),
		CodeHash:  randBytes(32),
		Data:      randBytes(128), // parse it
		DataHash:  randBytes(32),

		Types: []string{"item"},
	}

	accNoState = core.AccountState{
		Latest:     true,
		Address:    randAddr(),
		IsActive:   false,
		Status:     core.NonExist,
		Balance:    100e9,
		LastTxLT:   randLT(),
		LastTxHash: randBytes(32),
	}

	accDataWallet = core.AccountData{
		Address:    accWallet.Address,
		LastTxLT:   accWallet.LastTxLT,
		LastTxHash: accWallet.LastTxHash,
		Types:      accWallet.Types,
	}

	accDataItem = core.AccountData{
		Address:      accItem.Address,
		LastTxLT:     accItem.LastTxLT,
		LastTxHash:   accItem.LastTxHash,
		Types:        accItem.Types,
		OwnerAddress: randAddr(),
		NFTCollectionData: core.NFTCollectionData{
			NextItemIndex: 43,
		},
		NFTRoyaltyData: core.NFTRoyaltyData{
			RoyaltyAddress: randAddr(),
		},
		NFTContentData: core.NFTContentData{
			ContentURI: "git://asdf.t",
		},
		NFTItemData: core.NFTItemData{
			ItemIndex:         42,
			CollectionAddress: randAddr(),
		},
	}

	msgExtWallet = core.Message{
		Type:          core.ExternalIn,
		Hash:          randBytes(32),
		DstAddress:    accWallet.Address,
		Body:          randBytes(128),
		BodyHash:      randBytes(32),
		StateInitCode: nil,
		StateInitData: nil,
		CreatedAt:     randTs(),
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

		TotalFees:   1e5,
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

		Amount: 1e5,

		IHRDisabled: false,
		IHRFee:      0,
		FwdFee:      0,

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

		TotalFees: 1e3,

		StateUpdate: randBytes(32),
		Description: randBytes(32),

		OrigStatus: core.Active,
		EndStatus:  core.Active,

		CreatedAt: msgOutWallet.CreatedAt,
		CreatedLT: accItem.LastTxLT, // msgInItem.CreatedLT + 1,
	}

	msgInItemPayload = core.MessagePayload{
		Type: core.Internal,
		Hash: msgOutWallet.Hash,

		SrcAddress:  msgOutWallet.SrcAddress,
		SrcContract: abi.ContractName(accWallet.Types[0]),
		DstAddress:  msgOutWallet.DstAddress,
		DstContract: abi.ContractName(accItem.Types[0]),

		BodyHash:      msgOutWallet.BodyHash,
		OperationID:   msgOutWallet.OperationID,
		OperationName: "item_transfer",
		DataJSON:      "{\"new_owner\": \"kkkkkk\"}",

		CreatedAt: msgOutWallet.CreatedAt,
		CreatedLT: msgOutWallet.CreatedLT,
	}
)
