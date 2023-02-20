# Basic ton structures

## Block

```go
package tlb

// LookupBlock(workchain, shard, seq) => *tlb.BlockInfo
type BlockInfo struct {
    Workchain int32
    Shard     int64
    SeqNo     uint32
    RootHash  []byte
    FileHash  []byte
}
// GetBlockShardsInfo(master *tlb.BlockInfo) => shards []*tlb.BlockInfo
```

## Transaction

```go
package tlb

// GetBlockTransactions(block *tlb.BlockInfo) => (addr *address.Address, lt uint64)
// GetTransaction(block *tlb.BlockInfo, addr *address.Address, lt uint64) => *tlb.Transaction
type Transaction struct {
    Hash        []byte
    AccountAddr []byte
    LT          uint64
    PrevTxHash  []byte
    PrevTxLT    uint64
    Now         uint32
    OrigStatus  AccountStatus
    EndStatus   AccountStatus
    IO          struct {
        In      *Message 
        Out     []*Message
    }
    TotalFees   Coins
    StateUpdate *cell.Cell
    Description *cell.Cell
}
```

## Message

```go
package tlb

// *tlb.Transaction => (in *tlb.Message, out []*tlb.Message)
type AnyMessage struct {
    Type            MessageType        // external in, external out, internal
    
    Incoming        bool
    SrcAddr         *Address
    DstAddr         *Address
    Amount          Coins              // internal

    Bounce          bool               // internal
    Bounced         bool               // internal
    IHRDisabled     bool               // internal
    IHRFee          Coins              // internal
    FwdFee          Coins              // internal
    ImportFee       Coins              // external
    ExtraCurrencies *Dictionary        // internal

    StateInit       *StateInit
    Body            *cell.Cell
    
    CreatedLT       uint64
    CreatedAt       uint32
}
```

## Account

```go
package tlb

const (
    AccountStatusActive   = "ACTIVE"
    AccountStatusUninit   = "UNINIT"
    AccountStatusFrozen   = "FROZEN"
    AccountStatusNonExist = "NON_EXIST"
)

// GetAccount(block *tlb.BlockInfo, addr *address.Address) => *tlb.Account
type Account struct {
    IsActive                    bool
    State struct {
        Address                 *Address
        IsValid                 bool
        Status                  AccountStatus
        Balance                 Coins
        StateHash               []byte
        StateInit struct {
            Depth               uint64
            TickTock struct {
                Tick            bool
                Tock            bool
            }
            Code                *cell.Cell
            Data                *cell.Cell
            Lib                 *cell.Dictionary
        }
        StorageInfo struct {
            StorageUsed struct {
                BitsUsed        uint64
                CellsUsed       uint64
                PublicCellsUsed uint64
            }
            LastPaid            uint32
            DuePayment          *big.Int
        }
        LastTxLT uint64
    }
    Data                        *cell.Cell
    Code                        *cell.Cell
    LastTxLT                    uint64
    LastTxHash                  []byte
}
```

## Higher-level definitions

### Account data parsing

```go
package core

type ContractInterface struct {
    Name       ContractType // nft item/collection, jetton wallet/minter, telemint, etc.
    Address    string       // [optional] match contract type by address
    Code       []byte       // [optional] match contract type by code cells
    GetMethods []string     // [optional] match by presence of get methods in contract
    // We cannot define schema to parse return values of get methods.
    // But we can define schema for data cell.
}
```

### Message data parsing

```go
package core

type ContractOperation struct {
    Name         string       // nft collection mint/change_owner, nft item transfer/change_content
    ContractName ContractType // match message by contract type of src/dst addr
    Outgoing     bool         // if operation is going from contract
    OperationID  uint32       // match contract operation by operation id
    Schema       string       // json-encoded schema of message body
}
```

# Known contracts

## [ton-blockchain/token-contract/nft](https://github.com/ton-blockchain/token-contract/tree/1ad314a98d20b41241d5329e1786fc894ad811de/nft)

### NFT collection

Incoming messages:
1. `op::get_royalty_params() == 0x693d3950`
2. `(op == 1) ;; deploy new nft`
3. `(op == 2) ;; batch deploy of new nfts`
4. `(op == 3) ;; change owner`
5. `(op == 4) ;; change content, if editable`

Outgoing messages:
1. `op::report_royalty_params() == 0xa8cb00ad`

Get methods:
1. `(int, cell, slice) get_collection_data()`
2. `slice get_nft_address_by_index(int index)`
3. `(int, int, slice) royalty_params()`
4. `cell get_nft_content(int index, cell individual_nft_content)`
5. `slice get_editor()`

### NFT item

Incoming messages:
1. `op::transfer() == 0x5fcc3d14`
2. `op::get_static_data() == 0x2fcb26a2`
3. `op::edit_content() == 0x1a0b9d51 ;; if editable`
4. `op::transfer_editorship() == 0x1c04412a`

Outgoing messages:
1. `op::ownership_assigned() == 0x05138d91`
2. `op::excesses() == 0xd53276db`
3. `op::report_static_data() == 0x8b771735`
4. `op::editorship_assigned() == 0x511a4463`

Get methods:
1. `(int, int, slice, slice, cell) get_nft_data() ;; (init?, index, collection_address, owner_address, content)`
2. `slice get_editor()`

### NFT sale

Hard to parse it as every marketplace has its own custom NFT sale contracts.

Operations:
1. `(op == 1) ;; just accept coins`
2. `(op == 2) ;; buy`
3. `(op == 3) ;; cancel sale`
4. `op::ownership_assigned() ;; on creation of nft sale by nft item`

Get methods:
1. `(slice, slice, slice, int, int, slice, int) get_sale_data() ;; marketplace_address, nft_address, nft_owner_address, full_price, marketplace_fee, royalty_address, royalty_amount`

## [ton-blockchain/token-contract/ft](https://github.com/ton-blockchain/token-contract/tree/1ad314a98d20b41241d5329e1786fc894ad811de/ft)

### Jetton minter

Operations:

1. `op::mint()`
2. `op::burn_notification()`
3. ICO mint on empty body (buy jetton for TON coins)

Get methods:

1. `(int, int, slice, cell, cell) get_jetton_data() ;; (total_supply, -1, admin_address, content, jetton_wallet_code)`
2. `slice get_wallet_address(slice owner_address)`

### Jetton wallet

Data load:

`(int balance, slice owner_address, slice jetton_master_address, cell jetton_wallet_code) = load_data();`

Get methods:

1. `(int, slice, slice, cell) get_wallet_data() ;; balance, owner_address, jetton_master_address, jetton_wallet_code`

Operations:

1. `op::transfer()) ;; outgoing transfer`
2. `op::internal_transfer()) ;; incoming transfer`
3. `op::burn()) ;; burn`
4. `op::transfer_notification()` -- outgoing message to owner

## [TelegramMessenger/telemint](https://github.com/TelegramMessenger/telemint/tree/main/func)

### NFT collection

Operations:

1. `op::telemint_msg_deploy`
2. `op::teleitem_msg_deploy` -- outgoing message to item

Get methods:

1. `(int, cell, slice) get_collection_data()`
2. `slice get_full_domain()`
3. `slice get_nft_address_by_index(int index)`
4. `cell get_nft_content(int index, cell individual_nft_content)`
5. `(int, cell) dnsresolve(slice subdomain, int category)`

### NFT item

Operations:

1. `op::teleitem_msg_deploy` -- incoming message from collection
2. `op::teleitem_start_auction`
3. `op::teleitem_cancel_auction`
4. `op::nft_cmd_transfer`
5. `op::change_dns_record`
6. `op::get_royalty_params`
7. `op::nft_cmd_get_static_data`
8. process new bid: `op == 0`
9. `op::fill_up` -- outgoing message on `send_money` method
10. `op::ownership_assigned` -- outgoing message on NFT transfer
11. `op::outbid_notification` -- outgoing message to previous bidder on outbid
12. `op::teleitem_ok` -- success on start/cancel auction or change of dns record

Get methods:

1. `(int, int, slice, slice, cell) get_nft_data()`
2. `slice get_full_domain()`
3. `slice get_telemint_token_name()`
4. `(slice, int, int, int, int) get_telemint_auction_state()`
5. `(slice, int, int, int, int, int) get_telemint_auction_config()`
6. `(int, int, slice) royalty_params()`
7. `(int, cell) dnsresolve(slice subdomain, int category)`

## [ton-blockchain/dns-contract](https://github.com/ton-blockchain/dns-contract/tree/main/func)

TODO

## [getgems-io/nft-contracts](https://github.com/getgems-io/nft-contracts/blob/main/packages/contracts/sources)

TODO

## TONScan labels

[Link](https://raw.githubusercontent.com/menschee/tonscanplus/main/data.json)

# ClickHouse query examples

Get amount of unique wallets with given versions:

```clickhouse
SELECT count() FROM (
    SELECT any(address)
    FROM accounts
    WHERE hasAny(types, ['wallet_V3R1', 'wallet_V3R2', 'wallet_V4R1', 'wallet_V4R2'])
    GROUP BY address
)
```

Select rich wallets with balance more than 100k TON:

```clickhouse
SELECT any(address), max(balance) as b, any(types)
FROM accounts
WHERE hasAny(types, ['wallet_V3R1', 'wallet_V3R2', 'wallet_V4R1', 'wallet_V4R2'])
    AND balance > 100 * 1e3 * 1e9
GROUP BY address
ORDER BY b DESC
```

Select TOP 25 rich accounts, theirs types and balance:

```clickhouse
SELECT any(address), any(types), max(ton) FROM (
    SELECT address, types, balance / 1e9 AS ton
    FROM accounts
    ORDER BY balance DESC
    LIMIT 100
) GROUP BY address ORDER BY max(ton) DESC LIMIT 25
```

Select NFT items data:

```clickhouse
SELECT
    address,
    types,
    owner_address,
    content_uri,
    item_index,
    collection_address
FROM account_data
WHERE hasAny(types, ['nft_item', 'nft_item_editable'])
LIMIT 50
```

Get all nft sales with price more than 5 TON:

```clickhouse
SELECT hex(tx_hash), src_addr, dst_addr, amount / 1e9 AS ton
FROM messages
WHERE dst_addr IN (
    SELECT any(address)
    FROM accounts
    WHERE has(types, 'nft_sale')
    GROUP BY address
) AND ton > 5
```

Show all NFT numbers and all owners:

```clickhouse
SELECT
    any(address),
    any(content_uri),
    groupArray(owner_address)
FROM account_data
WHERE hasAny(types, ['nft_item']) AND (collection_address = 'EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N')
GROUP BY address
```