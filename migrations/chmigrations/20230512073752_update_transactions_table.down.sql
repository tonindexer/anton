ALTER TABLE transactions ADD COLUMN state_update String;

--migration:split

ALTER TABLE transactions DROP COLUMN result_code;
