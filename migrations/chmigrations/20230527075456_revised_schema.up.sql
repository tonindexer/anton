--
-- Clickhouse SHOW CREATE TABLE
--


CREATE TABLE account_states
(
    address String,
    workchain Int32,
    shard Int64,
    block_seq_no UInt32,
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
    types Array(String),
    owner_address String,
    minter_address String,
    content_uri String,
    content_name String,
    content_description String,
    content_image String,
    content_image_data String,
    jetton_balance UInt256,
    updated_at DateTime
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(updated_at)
ORDER BY (address, last_tx_lt)
SETTINGS index_granularity = 8192;


--migration:split


CREATE TABLE messages
(
    type LowCardinality(String),
    hash String,
    src_address String,
    src_tx_lt UInt64,
    src_workchain Int32,
    src_shard Int64,
    src_block_seq_no UInt32,
    dst_address String,
    dst_tx_lt UInt64,
    dst_workchain Int32,
    dst_shard Int64,
    dst_block_seq_no UInt32,
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
    src_contract LowCardinality(String),
    dst_contract LowCardinality(String),
    operation_name LowCardinality(String),
    error String,
    created_at DateTime,
    created_lt UInt64
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY hash
SETTINGS index_granularity = 8192;


--migration:split


CREATE TABLE transactions
(
    address String,
    hash String,
    workchain Int32,
    shard Int64,
    block_seq_no UInt32,
    prev_tx_hash String,
    prev_tx_lt UInt64,
    in_msg_hash String,
    in_amount UInt256,
    out_msg_count UInt16,
    out_amount UInt256,
    total_fees UInt256,
    description String,
    compute_phase_exit_code Int32,
    action_phase_result_code Int32,
    orig_status LowCardinality(String),
    end_status LowCardinality(String),
    created_at DateTime,
    created_lt UInt64
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY (address, created_lt)
SETTINGS index_granularity = 8192;


--migration:split


CREATE TABLE block_info
(
    workchain Int32,
    shard Int64,
    seq_no UInt32,
    file_hash String,
    root_hash String,
    scanned_at DateTime
)
ENGINE = ReplacingMergeTree
ORDER BY (workchain, shard, seq_no)
SETTINGS index_granularity = 8192;
