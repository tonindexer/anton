SET statement_timeout = 0;

--bun:split

-- add new interfaces from `abi/known` directory
DELETE FROM contract_interfaces;
DELETE FROM contract_operations;

--bun:split

ALTER TABLE contract_interfaces DROP COLUMN code_hash;

--bun:split

ALTER TABLE contract_interfaces DROP COLUMN get_methods;

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN get_methods_desc text;

--bun:split

ALTER TABLE ONLY contract_interfaces
    DROP CONSTRAINT contract_interfaces_get_method_hashes_key;

ALTER TABLE ONLY contract_interfaces
    ADD CONSTRAINT contract_interfaces_addresses_key UNIQUE (addresses);

CREATE UNIQUE INDEX contract_interfaces_get_method_hashes_idx ON contract_interfaces (get_method_hashes)
    WHERE addresses IS NULL and code IS NULL;
