--
-- PostgreSQL database dump
--

SET statement_timeout = 0;


CREATE TABLE block_info (
    workchain integer NOT NULL,
    shard bigint NOT NULL,
    seq_no integer NOT NULL,
    file_hash bytea NOT NULL,
    root_hash bytea NOT NULL,
    master_workchain integer,
    master_shard bigint,
    master_seq_no integer,
    scanned_at timestamp without time zone NOT NULL,
    CONSTRAINT block_info_pkey PRIMARY KEY (workchain, shard, seq_no),
    CONSTRAINT block_info_file_hash_key UNIQUE (file_hash),
    CONSTRAINT block_info_root_hash_key UNIQUE (root_hash)
);


--bun:split
CREATE TYPE label_category AS ENUM (
    'centralized_exchange',
    'scam'
);

--bun:split
CREATE TABLE address_labels (
    address bytea NOT NULL,
    name text,
    categories label_category[],
    CONSTRAINT address_labels_pkey PRIMARY KEY (address)
);


--bun:split
CREATE TYPE account_status AS ENUM (
    'UNINIT',
    'ACTIVE',
    'FROZEN',
    'NON_EXIST'
);

--bun:split
CREATE TABLE account_states (
    address bytea NOT NULL,
    workchain integer NOT NULL,
    shard bigint NOT NULL,
    block_seq_no integer NOT NULL,
    is_active boolean,
    status account_status,
    balance numeric,
    last_tx_lt bigint NOT NULL,
    last_tx_hash bytea NOT NULL,
    state_hash bytea,
    code bytea,
    code_hash bytea,
    data bytea,
    data_hash bytea,
    get_method_hashes integer[],
    types text[],
    owner_address bytea,
    minter_address bytea,
    executed_get_methods jsonb,
    content_uri character varying,
    content_name character varying,
    content_description character varying,
    content_image character varying,
    content_image_data bytea,
    jetton_balance numeric,
    updated_at timestamp without time zone NOT NULL,
    CONSTRAINT account_states_pkey PRIMARY KEY (address, last_tx_lt),
    CONSTRAINT account_states_last_tx_hash_key UNIQUE (last_tx_hash)
);

--bun:split
CREATE TABLE latest_account_states (
    address bytea NOT NULL,
    last_tx_lt bigint NOT NULL,
    CONSTRAINT latest_account_states_pkey PRIMARY KEY (address),
    CONSTRAINT latest_account_states_address_last_tx_lt_fkey FOREIGN KEY (address, last_tx_lt) REFERENCES account_states(address, last_tx_lt)
);


--bun:split
CREATE TYPE message_type AS ENUM (
    'EXTERNAL_IN',
    'EXTERNAL_OUT',
    'INTERNAL'
);

--bun:split
CREATE TABLE messages (
    type message_type NOT NULL,
    hash bytea NOT NULL,
    src_address bytea,
    src_tx_lt bigint,
    src_workchain integer NOT NULL,
    src_shard bigint NOT NULL,
    src_block_seq_no integer NOT NULL,
    dst_address bytea,
    dst_tx_lt bigint,
    dst_workchain integer NOT NULL,
    dst_shard bigint NOT NULL,
    dst_block_seq_no integer NOT NULL,
    bounce boolean NOT NULL,
    bounced boolean NOT NULL,
    amount numeric,
    ihr_disabled boolean NOT NULL,
    ihr_fee numeric,
    fwd_fee numeric,
    body bytea,
    body_hash bytea,
    operation_id integer,
    transfer_comment character varying,
    state_init_code bytea,
    state_init_data bytea,
    src_contract character varying,
    dst_contract character varying,
    operation_name character varying,
    data_json jsonb,
    error character varying,
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL,
    CONSTRAINT messages_pkey PRIMARY KEY (hash),
    CONSTRAINT messages_tx_lt_notnull CHECK ((
        (
            (type = 'EXTERNAL_OUT') AND (src_address IS NOT NULL) AND (src_tx_lt IS NOT NULL) AND (dst_address IS NULL) AND (dst_tx_lt IS NULL)
        ) OR (
            (type = 'EXTERNAL_IN') AND (src_address IS NULL) AND (src_tx_lt IS NULL) AND (dst_address IS NOT NULL) AND (dst_tx_lt IS NOT NULL)
        ) OR (
            (type = 'INTERNAL') AND
            (src_workchain != -1 OR dst_workchain != -1) AND
            -- (src_address IS NOT NULL) AND -- counterexample in testnet: cf86b1d575e0a34720abc911edc6ce303b2f8cf842117bafb0e18af709ca3b17
            (src_tx_lt IS NOT NULL) AND
            -- (dst_address IS NOT NULL) AND -- destination can be null (burn address)
            (dst_tx_lt IS NOT NULL)
        ) OR (
            (type = 'INTERNAL') AND src_workchain = -1 AND dst_workchain = -1
        )
    ))
);

--bun:split
CREATE TABLE transactions (
    address bytea,
    hash bytea NOT NULL,
    workchain integer NOT NULL,
    shard bigint NOT NULL,
    block_seq_no integer NOT NULL,
    prev_tx_hash bytea,
    prev_tx_lt bigint,
    in_msg_hash bytea,
    in_amount numeric NOT NULL,
    out_msg_count smallint NOT NULL,
    out_amount numeric NOT NULL,
    total_fees numeric,
    description bytea NOT NULL,
    compute_phase_exit_code integer NOT NULL,
    action_phase_result_code integer NOT NULL,
    orig_status account_status NOT NULL,
    end_status account_status NOT NULL,
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL,
    CONSTRAINT transactions_pkey PRIMARY KEY (hash)
);


--bun:split
CREATE TABLE contract_interfaces (
    name character varying NOT NULL,
    addresses bytea[],
    code bytea,
    get_methods_desc text,
    get_method_hashes integer[],
    CONSTRAINT contract_interfaces_pkey PRIMARY KEY (name),
    CONSTRAINT contract_interfaces_addresses_key UNIQUE (addresses),
    CONSTRAINT contract_interfaces_code_key UNIQUE (code)
);

--bun:split
CREATE TABLE contract_operations (
    operation_name character varying,
    contract_name character varying NOT NULL,
    message_type message_type NOT NULL,
    outgoing boolean NOT NULL,
    operation_id integer NOT NULL,
    schema jsonb,
    CONSTRAINT contract_operations_pkey PRIMARY KEY (contract_name, outgoing, operation_id),
    CONSTRAINT contract_interfaces_uniq_name UNIQUE (operation_name, contract_name)
);


--bun:split
CREATE INDEX account_states_address_idx ON account_states USING hash (address);

--bun:split
CREATE UNIQUE INDEX account_states_address_workchain_shard_block_seq_no_idx ON account_states USING btree (address, workchain, shard, block_seq_no);

--bun:split
CREATE INDEX account_states_last_tx_lt_idx ON account_states USING btree (last_tx_lt);

--bun:split
CREATE INDEX account_states_minter_address_idx ON account_states USING hash (minter_address) WHERE (minter_address IS NOT NULL);

--bun:split
CREATE INDEX account_states_owner_address_idx ON account_states USING hash (owner_address) WHERE (owner_address IS NOT NULL);

--bun:split
CREATE INDEX account_states_types_idx ON account_states USING gin (types);


--bun:split
CREATE UNIQUE INDEX contract_interfaces_get_method_hashes_idx ON contract_interfaces USING btree (get_method_hashes) WHERE ((addresses IS NULL) AND (code IS NULL));

--bun:split
CREATE INDEX latest_account_states_last_tx_lt_idx ON latest_account_states USING btree (last_tx_lt);


--bun:split
CREATE INDEX messages_created_lt_idx ON messages USING btree (created_lt);

--bun:split
CREATE INDEX messages_dst_address_idx ON messages USING hash (dst_address) WHERE (dst_address IS NOT NULL);

--bun:split
CREATE INDEX messages_src_address_idx ON messages USING hash (src_address) WHERE (src_address IS NOT NULL);

--bun:split
CREATE INDEX messages_src_contract_idx ON messages USING hash (src_contract) WHERE (src_contract IS NOT NULL);

--bun:split
CREATE INDEX messages_dst_contract_idx ON messages USING hash (dst_contract) WHERE (src_contract IS NOT NULL);

--bun:split
CREATE INDEX messages_operation_id_idx ON messages USING hash (operation_id);

--bun:split
CREATE INDEX messages_operation_name_idx ON messages USING hash (operation_name) WHERE (operation_name IS NOT NULL);

--bun:split
CREATE UNIQUE INDEX messages_src_address_created_lt_idx ON messages USING btree (src_address, created_lt) WHERE (src_address IS NOT NULL);

--bun:split
CREATE INDEX messages_src_address_src_tx_lt_idx ON messages USING btree (src_address, src_tx_lt) WHERE (src_address IS NOT NULL);


--bun:split
CREATE UNIQUE INDEX transactions_address_created_lt_idx ON transactions USING btree (address, created_lt);

--bun:split
CREATE INDEX transactions_address_idx ON transactions USING hash (address);

--bun:split
CREATE INDEX transactions_created_lt_idx ON transactions USING btree (created_lt);

--bun:split
CREATE INDEX transactions_in_msg_hash_idx ON transactions USING hash (in_msg_hash);

--bun:split
CREATE INDEX transactions_workchain_shard_block_seq_no_idx ON transactions USING btree (workchain, shard, block_seq_no);
