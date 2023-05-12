ALTER TABLE transactions ADD COLUMN state_update bytea;

--bun:split

ALTER TABLE transactions DROP COLUMN result_code;