SET statement_timeout = 0;

BEGIN;
    ALTER TABLE latest_account_states ALTER COLUMN created_lt SET NOT NULL;

    CREATE INDEX latest_account_states_created_lt_idx ON latest_account_states USING btree (created_lt);
COMMIT;
