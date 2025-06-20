--bun:split
CREATE TABLE latest_parsed_account_states (
    address bytea NOT NULL,
    last_tx_lt bigint NOT NULL,
    types text[],
    owner_address bytea,
    minter_address bytea,
    CONSTRAINT latest_parsed_account_states_pkey PRIMARY KEY (address),
    CONSTRAINT latest_parsed_account_states_address_last_tx_lt_fkey FOREIGN KEY (address, last_tx_lt) REFERENCES account_states(address, last_tx_lt)
);

-- -- migrate to the new table
-- INSERT INTO latest_parsed_account_states (address, last_tx_lt, types, owner_address, minter_address)
-- SELECT ls.address, ls.last_tx_lt, types, owner_address, minter_address
-- FROM latest_account_states ls
-- INNER JOIN account_states s ON ls.address = s.address AND ls.last_tx_lt = s.last_tx_lt
-- WHERE array_length(s.types, 1) > 0 OR owner_address IS NOT NULL OR minter_address IS NOT NULL
-- ORDER BY s.last_tx_lt
-- LIMIT 10000;

--bun:split
CREATE INDEX latest_parsed_account_states_types_idx ON latest_parsed_account_states USING gin (types);

--bun:split
CREATE INDEX latest_parsed_account_states_minter_address_idx ON latest_parsed_account_states USING btree (minter_address) WHERE (minter_address IS NOT NULL);

--bun:split
CREATE INDEX latest_parsed_account_states_owner_address_idx ON latest_parsed_account_states USING btree (owner_address) WHERE (owner_address IS NOT NULL);
