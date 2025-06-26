-- DROP PROCEDURE batch_update_latest_parsed_account_states;

BEGIN;

    ALTER TABLE latest_account_states
        DROP COLUMN types,
        DROP COLUMN owner_address,
        DROP COLUMN minter_address;

COMMIT;
