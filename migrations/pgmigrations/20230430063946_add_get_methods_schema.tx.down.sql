SET statement_timeout = 0;

--bun:split

ALTER TABLE contract_interfaces DROP COLUMN get_methods_desc;

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN get_methods text[];

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN code_hash bytea;

--bun:split

DROP INDEX contract_interfaces_get_method_hashes_idx;

ALTER TABLE ONLY contract_interfaces
    DROP CONSTRAINT contract_interfaces_addresses_key;

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_get_method_hashes_key UNIQUE (get_method_hashes);
