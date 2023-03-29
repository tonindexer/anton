--
-- PostgreSQL database dump
--

-- Dumped from database version 15.1 (Debian 15.1-1.pgdg110+1)
-- Dumped by pg_dump version 15.1 (Debian 15.1-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: account_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.account_status AS ENUM (
    'UNINIT',
    'ACTIVE',
    'FROZEN',
    'NON_EXIST'
);


--
-- Name: message_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.message_type AS ENUM (
    'EXTERNAL_IN',
    'EXTERNAL_OUT',
    'INTERNAL'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account_data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.account_data (
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


--
-- Name: account_states; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.account_states (
    address bytea NOT NULL,
    is_active boolean,
    status public.account_status,
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


--
-- Name: block_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.block_info (
    workchain integer NOT NULL,
    shard bigint NOT NULL,
    seq_no integer NOT NULL,
    file_hash bytea NOT NULL,
    root_hash bytea NOT NULL,
    master_workchain integer,
    master_shard bigint,
    master_seq_no integer
);


--
-- Name: contract_interfaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contract_interfaces (
    name character varying NOT NULL,
    addresses bytea[],
    code bytea,
    code_hash bytea,
    get_methods character varying[],
    get_method_hashes integer[]
);


--
-- Name: contract_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contract_operations (
    name character varying,
    contract_name character varying NOT NULL,
    outgoing boolean NOT NULL,
    operation_id integer NOT NULL,
    schema jsonb
);


--
-- Name: latest_account_states; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.latest_account_states (
    address bytea NOT NULL,
    last_tx_lt bigint NOT NULL
);


--
-- Name: message_payloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message_payloads (
    type public.message_type NOT NULL,
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


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    type public.message_type NOT NULL,
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


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    hash bytea NOT NULL,
    address bytea NOT NULL,
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
    orig_status public.account_status NOT NULL,
    end_status public.account_status NOT NULL,
    created_at timestamp without time zone NOT NULL,
    created_lt bigint NOT NULL
);


--
-- Name: account_data account_data_last_tx_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_data
    ADD CONSTRAINT account_data_last_tx_hash_key UNIQUE (last_tx_hash);


--
-- Name: account_data account_data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_data
    ADD CONSTRAINT account_data_pkey PRIMARY KEY (address, last_tx_lt);


--
-- Name: account_states account_states_last_tx_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_states
    ADD CONSTRAINT account_states_last_tx_hash_key UNIQUE (last_tx_hash);


--
-- Name: account_states account_states_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_states
    ADD CONSTRAINT account_states_pkey PRIMARY KEY (address, last_tx_lt);


--
-- Name: block_info block_info_file_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_info
    ADD CONSTRAINT block_info_file_hash_key UNIQUE (file_hash);


--
-- Name: block_info block_info_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_info
    ADD CONSTRAINT block_info_pkey PRIMARY KEY (workchain, shard, seq_no);


--
-- Name: block_info block_info_root_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.block_info
    ADD CONSTRAINT block_info_root_hash_key UNIQUE (root_hash);


--
-- Name: contract_interfaces contract_interfaces_code_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_interfaces
    ADD CONSTRAINT contract_interfaces_code_hash_key UNIQUE (code_hash);


--
-- Name: contract_interfaces contract_interfaces_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_interfaces
    ADD CONSTRAINT contract_interfaces_code_key UNIQUE (code);


--
-- Name: contract_interfaces contract_interfaces_get_method_hashes_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_interfaces
    ADD CONSTRAINT contract_interfaces_get_method_hashes_key UNIQUE (get_method_hashes);


--
-- Name: contract_interfaces contract_interfaces_get_methods_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_interfaces
    ADD CONSTRAINT contract_interfaces_get_methods_key UNIQUE (get_methods);


--
-- Name: contract_interfaces contract_interfaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_interfaces
    ADD CONSTRAINT contract_interfaces_pkey PRIMARY KEY (name);


--
-- Name: contract_operations contract_operations_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_operations
    ADD CONSTRAINT contract_operations_name_key UNIQUE (name);


--
-- Name: contract_operations contract_operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contract_operations
    ADD CONSTRAINT contract_operations_pkey PRIMARY KEY (contract_name, outgoing, operation_id);


--
-- Name: latest_account_states latest_account_states_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.latest_account_states
    ADD CONSTRAINT latest_account_states_pkey PRIMARY KEY (address);


--
-- Name: message_payloads message_payloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_payloads
    ADD CONSTRAINT message_payloads_pkey PRIMARY KEY (hash);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (hash);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (hash);


--
-- Name: account_data_minter_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX account_data_minter_address_idx ON public.account_data USING hash (minter_address) WHERE (length(minter_address) > 0);


--
-- Name: account_data_owner_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX account_data_owner_address_idx ON public.account_data USING hash (owner_address) WHERE (length(owner_address) > 0);


--
-- Name: account_data_types_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX account_data_types_idx ON public.account_data USING gin (types);


--
-- Name: account_states_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX account_states_address_idx ON public.account_states USING hash (address);


--
-- Name: account_states_last_tx_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX account_states_last_tx_lt_idx ON public.account_states USING btree (last_tx_lt);


--
-- Name: block_info_workchain_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX block_info_workchain_idx ON public.block_info USING hash (workchain);


--
-- Name: latest_account_states_last_tx_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX latest_account_states_last_tx_lt_idx ON public.latest_account_states USING btree (last_tx_lt);


--
-- Name: message_payloads_dst_contract_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX message_payloads_dst_contract_idx ON public.message_payloads USING hash (dst_contract) WHERE (length((dst_contract)::text) > 0);


--
-- Name: message_payloads_minter_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX message_payloads_minter_address_idx ON public.message_payloads USING hash (minter_address) WHERE (length(minter_address) > 0);


--
-- Name: message_payloads_operation_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX message_payloads_operation_name_idx ON public.message_payloads USING hash (operation_name);


--
-- Name: message_payloads_src_contract_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX message_payloads_src_contract_idx ON public.message_payloads USING hash (src_contract) WHERE (length((src_contract)::text) > 0);


--
-- Name: messages_created_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_created_lt_idx ON public.messages USING btree (created_lt);


--
-- Name: messages_dst_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_dst_address_idx ON public.messages USING hash (dst_address) WHERE (length(dst_address) > 0);


--
-- Name: messages_src_address_created_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX messages_src_address_created_lt_idx ON public.messages USING btree (src_address, created_lt) WHERE (length(src_address) > 0);


--
-- Name: messages_src_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_src_address_idx ON public.messages USING hash (src_address) WHERE (length(src_address) > 0);


--
-- Name: messages_src_address_source_tx_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX messages_src_address_source_tx_lt_idx ON public.messages USING btree (src_address, source_tx_lt) WHERE ((length(src_address) > 0) AND (source_tx_lt > 0));


--
-- Name: transactions_address_created_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX transactions_address_created_lt_idx ON public.transactions USING btree (address, created_lt);


--
-- Name: transactions_address_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_address_idx ON public.transactions USING hash (address);


--
-- Name: transactions_block_workchain_block_shard_block_seq_no_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_block_workchain_block_shard_block_seq_no_idx ON public.transactions USING btree (block_workchain, block_shard, block_seq_no);


--
-- Name: transactions_created_lt_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_created_lt_idx ON public.transactions USING btree (created_lt);


--
-- Name: transactions_in_msg_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_in_msg_hash_idx ON public.transactions USING hash (in_msg_hash);


--
-- Name: latest_account_states latest_account_states_address_last_tx_lt_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.latest_account_states
    ADD CONSTRAINT latest_account_states_address_last_tx_lt_fkey FOREIGN KEY (address, last_tx_lt) REFERENCES public.account_states(address, last_tx_lt);


--
-- PostgreSQL database dump complete
--
