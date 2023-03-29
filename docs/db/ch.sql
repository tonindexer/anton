--
-- Clickhouse SHOW CREATE TABLE
--


CREATE TABLE block_info
(
    workchain Int32,
    shard Int64,
    seq_no UInt32,
    file_hash String,
    root_hash String
)
ENGINE = ReplacingMergeTree
PARTITION BY workchain
ORDER BY (workchain, shard, seq_no);


CREATE TABLE account_data
(
    address String,
    last_tx_lt UInt64,
    last_tx_hash String,
    balance UInt256,
    types Array(String),
    owner_address String,
    minter_address String,
    next_item_index UInt256,
    royalty_address String,
    royalty_factor UInt16,
    royalty_base UInt16,
    content_uri String,
    content_name String,
    content_description String,
    content_image String,
    content_image_data String,
    initialized Bool,
    item_index UInt256,
    editor_address String,
    total_supply UInt256,
    mintable Bool,
    admin_address String,
    jetton_balance UInt256,
    errors Array(String),
    updated_at DateTime
)
ENGINE = ReplacingMergeTree
PARTITION BY types
ORDER BY (address, last_tx_lt);


CREATE TABLE account_states
(
    address String,
    is_active Bool,
    status LowCardinality(String),
    balance UInt256,
    last_tx_lt UInt64,
    last_tx_hash String,
    state_hash String,
    code String,
    code_hash String,
    data String,
    data_hash String,
    get_method_hashes Array(UInt32),
    updated_at DateTime
)
ENGINE = ReplacingMergeTree
PARTITION BY status
ORDER BY (address, last_tx_lt);


CREATE TABLE message_payloads
(
    type LowCardinality(String),
    hash String,
    src_address String,
    src_contract LowCardinality(String),
    dst_address String,
    dst_contract LowCardinality(String),
    amount UInt256,
    body_hash String,
    operation_id UInt32,
    operation_name LowCardinality(String),
    minter_address String,
    created_at DateTime,
    created_lt UInt64,
    error String
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY hash;


CREATE TABLE messages
(
    type LowCardinality(String),
    hash String,
    src_address String,
    dst_address String,
    source_tx_hash String,
    source_tx_lt UInt64,
    bounce Bool,
    bounced Bool,
    amount UInt256,
    ihr_disabled Bool,
    ihr_fee UInt256,
    fwd_fee UInt256,
    body String,
    body_hash String,
    operation_id UInt32,
    transfer_comment String,
    state_init_code String,
    state_init_data String,
    created_at DateTime,
    created_lt UInt64
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY hash;


CREATE TABLE transactions
(
    hash String,
    address String,
    block_workchain Int32,
    block_shard Int64,
    block_seq_no UInt32,
    prev_tx_hash String,
    prev_tx_lt UInt64,
    in_msg_hash String,
    in_amount UInt256,
    out_msg_count UInt16,
    out_amount UInt256,
    total_fees UInt256,
    state_update String,
    description String,
    orig_status LowCardinality(String),
    end_status LowCardinality(String),
    created_at DateTime,
    created_lt UInt64
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY (address, hash);