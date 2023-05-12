ALTER TABLE transactions DROP COLUMN state_update;

--migration:split

ALTER TABLE transactions ADD COLUMN result_code Int32;

--migration:split