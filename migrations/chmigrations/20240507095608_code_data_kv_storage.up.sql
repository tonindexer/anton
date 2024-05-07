CREATE TABLE account_states_code
(
    code_hash String,
    code String
)
ENGINE = EmbeddedRocksDB
PRIMARY KEY code_hash;

--migration:split

CREATE TABLE account_states_data
(
    data_hash String,
    data String
)
ENGINE = EmbeddedRocksDB
PRIMARY KEY data_hash;

--migration:split

-- INSERT INTO account_states_code SELECT code_hash, any(code) FROM account_states GROUP BY code_hash;
-- INSERT INTO account_states_data SELECT data_hash, any(data) FROM account_states GROUP BY data_hash;

-- ALTER TABLE account_states DROP COLUMN code;
-- ALTER TABLE account_states DROP COLUMN data;
