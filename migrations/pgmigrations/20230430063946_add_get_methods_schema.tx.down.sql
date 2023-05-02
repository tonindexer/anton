SET statement_timeout = 0;

--bun:split

ALTER TABLE contract_interfaces DROP COLUMN get_methods_desc;

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN get_methods text[];

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN code_hash bytea;
