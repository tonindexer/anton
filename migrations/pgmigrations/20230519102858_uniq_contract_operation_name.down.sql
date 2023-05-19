SET statement_timeout = 0;

--bun:split

ALTER TABLE ONLY contract_operations
    DROP CONSTRAINT contract_operations_name_contract_name_key;

--bun:split

ALTER TABLE ONLY contract_operations
    ADD CONSTRAINT contract_operations_name_key UNIQUE (name);
