--bun:split
DROP PROCEDURE batch_update_latest_parsed_account_states;


--bun:split
ALTER TABLE latest_account_states DROP COLUMN types;

--bun:split
ALTER TABLE latest_account_states DROP COLUMN owner_address;

--bun:split
ALTER TABLE latest_account_states DROP COLUMN minter_address;
