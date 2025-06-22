-- DROP PROCEDURE batch_fill_account_states_created_lt;

BEGIN;

    ALTER TABLE latest_account_states DROP COLUMN created_lt bigint;

COMMIT;
