SET statement_timeout = 0;

--bun:split

DELETE FROM contract_interfaces;
DELETE FROM contract_operations;

--bun:split

ALTER TABLE contract_interfaces DROP COLUMN get_methods;

--bun:split

ALTER TABLE contract_interfaces ADD COLUMN get_methods_desc jsonb;
