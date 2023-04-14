--
-- PostgreSQL database dump
--

SET statement_timeout = 0;

--bun:split

CREATE TYPE account_status AS ENUM (
    'UNINIT',
    'ACTIVE',
    'FROZEN',
    'NON_EXIST'
);

CREATE TABLE account_data (
    address bytea NOT NULL,
    last_tx_lt bigint NOT NULL,
    last_tx_hash bytea NOT NULL,
    balance numeric,
    types text[],
    owner_address bytea,
    minter_address bytea,
    next_item_index numeric,
    royalty_address bytea,
    royalty_factor integer,
    royalty_base integer,
    content_uri character varying,
    content_name character varying,
    content_description character varying,
    content_image character varying,
    content_image_data bytea,
    initialized boolean,
    item_index jsonb,
    editor_address bytea,
    total_supply numeric,
    mintable boolean,
    admin_address bytea,
    jetton_balance numeric,
    errors jsonb,
    updated_at timestamp without time zone NOT NULL
);

CREATE TABLE account_states (
    address bytea NOT NULL,
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
    updated_at timestamp without time zone NOT NULL
);

CREATE TABLE latest_account_states (
    address bytea NOT NULL,
    last_tx_lt bigint NOT NULL
);


--bun:split


CREATE TYPE message_type AS ENUM (
    'EXTERNAL_IN',
    'EXTERNAL_OUT',
    'INTERNAL'
);

CREATE TABLE message_payloads (
    type message_type NOT NULL,
    hash bytea NOT NULL,
    src_address bytea,
    src_contract character varying,
    dst_address bytea,
    dst_contract character varying,
    amount numeric,
    body_hash bytea NOT NULL,
    operation_id integer NOT NULL,
    operation_name character varying NOT NULL,
    data_json jsonb,
    minter_address bytea,
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL,
    error character varying
);

CREATE TABLE messages (
    type message_type NOT NULL,
    hash bytea NOT NULL,
    src_address bytea,
    dst_address bytea,
    source_tx_hash bytea,
    source_tx_lt bigint,
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
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL
);


CREATE TABLE transactions (
    address bytea NOT NULL,
    hash bytea NOT NULL,
    block_workchain integer NOT NULL,
    block_shard bigint NOT NULL,
    block_seq_no integer NOT NULL,
    prev_tx_hash bytea,
    prev_tx_lt bigint,
    in_msg_hash bytea,
    in_amount numeric NOT NULL,
    out_msg_count smallint NOT NULL,
    out_amount numeric NOT NULL,
    total_fees numeric,
    state_update bytea,
    description bytea,
    orig_status account_status NOT NULL,
    end_status account_status NOT NULL,
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL
);


CREATE TABLE block_info (
    workchain integer NOT NULL,
    shard bigint NOT NULL,
    seq_no integer NOT NULL,
    file_hash bytea NOT NULL,
    root_hash bytea NOT NULL,
    master_workchain integer,
    master_shard bigint,
    master_seq_no integer
);


CREATE TABLE contract_interfaces (
    name character varying NOT NULL,
    addresses bytea[],
    code bytea,
    code_hash bytea,
    get_methods character varying[],
    get_method_hashes integer[]
);

CREATE TABLE contract_operations (
    name character varying,
    contract_name character varying NOT NULL,
    outgoing boolean NOT NULL,
    operation_id integer NOT NULL,
    schema jsonb
);


-- constraints
--bun:split

ALTER TABLE ONLY account_data
    ADD CONSTRAINT account_data_last_tx_hash_key UNIQUE (last_tx_hash);

ALTER TABLE ONLY account_data
    ADD CONSTRAINT account_data_pkey PRIMARY KEY (address, last_tx_lt);


ALTER TABLE ONLY account_states
    ADD CONSTRAINT account_states_last_tx_hash_key UNIQUE (last_tx_hash);

ALTER TABLE ONLY account_states
    ADD CONSTRAINT account_states_pkey PRIMARY KEY (address, last_tx_lt);

ALTER TABLE ONLY latest_account_states
    ADD CONSTRAINT latest_account_states_pkey PRIMARY KEY (address);


ALTER TABLE ONLY block_info
    ADD CONSTRAINT block_info_file_hash_key UNIQUE (file_hash);

ALTER TABLE ONLY block_info
    ADD CONSTRAINT block_info_pkey PRIMARY KEY (workchain, shard, seq_no);

ALTER TABLE ONLY block_info
    ADD CONSTRAINT block_info_root_hash_key UNIQUE (root_hash);


ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_code_hash_key UNIQUE (code_hash);

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_code_key UNIQUE (code);

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_get_method_hashes_key UNIQUE (get_method_hashes);

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_get_methods_key UNIQUE (get_methods);

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_pkey PRIMARY KEY (name);


ALTER TABLE ONLY contract_operations
    ADD CONSTRAINT contract_operations_name_key UNIQUE (name);

ALTER TABLE ONLY contract_operations
    ADD CONSTRAINT contract_operations_pkey PRIMARY KEY (contract_name, outgoing, operation_id);


ALTER TABLE ONLY message_payloads
    ADD CONSTRAINT message_payloads_pkey PRIMARY KEY (hash);


ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (hash);


ALTER TABLE ONLY transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (hash);


-- indexes
--bun:split

CREATE INDEX account_data_minter_address_idx ON account_data USING hash (minter_address) WHERE (length(minter_address) > 0);

CREATE INDEX account_data_owner_address_idx ON account_data USING hash (owner_address) WHERE (length(owner_address) > 0);

CREATE INDEX account_data_types_idx ON account_data USING gin (types);


CREATE INDEX account_states_address_idx ON account_states USING hash (address);

CREATE INDEX account_states_last_tx_lt_idx ON account_states USING btree (last_tx_lt);


CREATE INDEX latest_account_states_last_tx_lt_idx ON latest_account_states USING btree (last_tx_lt);


CREATE INDEX block_info_workchain_idx ON block_info USING hash (workchain);


CREATE INDEX message_payloads_dst_contract_idx ON message_payloads USING hash (dst_contract) WHERE (length((dst_contract)::text) > 0);

CREATE INDEX message_payloads_minter_address_idx ON message_payloads USING hash (minter_address) WHERE (length(minter_address) > 0);

CREATE INDEX message_payloads_operation_name_idx ON message_payloads USING hash (operation_name);

CREATE INDEX message_payloads_src_contract_idx ON message_payloads USING hash (src_contract) WHERE (length((src_contract)::text) > 0);


CREATE INDEX messages_created_lt_idx ON messages USING btree (created_lt);

CREATE INDEX messages_dst_address_idx ON messages USING hash (dst_address) WHERE (length(dst_address) > 0);

CREATE UNIQUE INDEX messages_src_address_created_lt_idx ON messages USING btree (src_address, created_lt) WHERE (length(src_address) > 0);

CREATE INDEX messages_src_address_idx ON messages USING hash (src_address) WHERE (length(src_address) > 0);

CREATE INDEX messages_src_address_source_tx_lt_idx ON messages USING btree (src_address, source_tx_lt) WHERE ((length(src_address) > 0) AND (source_tx_lt > 0));


CREATE UNIQUE INDEX transactions_address_created_lt_idx ON transactions USING btree (address, created_lt);

CREATE INDEX transactions_address_idx ON transactions USING hash (address);

CREATE INDEX transactions_block_workchain_block_shard_block_seq_no_idx ON transactions USING btree (block_workchain, block_shard, block_seq_no);

CREATE INDEX transactions_created_lt_idx ON transactions USING btree (created_lt);

CREATE INDEX transactions_in_msg_hash_idx ON transactions USING hash (in_msg_hash);


-- foreign keys
--bun:split

ALTER TABLE ONLY latest_account_states
    ADD CONSTRAINT latest_account_states_address_last_tx_lt_fkey FOREIGN KEY (address, last_tx_lt) REFERENCES account_states(address, last_tx_lt);
