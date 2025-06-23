SET statement_timeout = 0;

BEGIN;
    DROP INDEX latest_account_states_created_lt_idx;

    ALTER TABLE latest_account_states ALTER COLUMN created_lt DROP NOT NULL;
COMMIT;
