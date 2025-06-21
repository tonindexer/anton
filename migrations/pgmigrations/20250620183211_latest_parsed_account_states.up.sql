--bun:split
ALTER TABLE latest_account_states ADD COLUMN types text[];

--bun:split
ALTER TABLE latest_account_states ADD COLUMN owner_address bytea;

--bun:split
ALTER TABLE latest_account_states ADD COLUMN minter_address bytea;


--bun:split
CREATE INDEX latest_account_states_types_idx ON latest_account_states USING gin (types);

--bun:split
CREATE INDEX latest_account_states_minter_address_idx ON latest_account_states USING btree (minter_address) WHERE (minter_address IS NOT NULL);

--bun:split
CREATE INDEX latest_account_states_owner_address_idx ON latest_account_states USING btree (owner_address) WHERE (owner_address IS NOT NULL);


-- --bun:split
-- CREATE OR REPLACE PROCEDURE batch_update_latest_account_states(
--     batch_size INT DEFAULT 10000,
--     start_from_lt BIGINT DEFAULT 0
-- )
-- LANGUAGE plpgsql
-- AS $$
-- DECLARE
-- last_processed_tx_lt BIGINT := start_from_lt;
--     rows_updated INT;
--     iteration_count INT := 0;
--     max_tx_lt BIGINT;
-- BEGIN
--     RAISE NOTICE 'Starting batch update process with batch size: %', batch_size;
--
--     LOOP
--         -- Update the next batch
--         WITH updated AS (
--             UPDATE latest_account_states las
--             SET
--                 types = s.types,
--                 owner_address = s.owner_address,
--                 minter_address = s.minter_address
--             FROM account_states s
--             WHERE las.address = s.address
--               AND las.last_tx_lt = s.last_tx_lt
--               AND s.last_tx_lt >= last_processed_tx_lt
--               AND (
--                   las.types IS DISTINCT FROM s.types OR
--                   las.owner_address IS DISTINCT FROM s.owner_address OR
--                   las.minter_address IS DISTINCT FROM s.minter_address
--               )
--             ORDER BY s.last_tx_lt
--             LIMIT batch_size
--             RETURNING s.last_tx_lt
--         )
--         SELECT COUNT(*), MAX(last_tx_lt) INTO rows_updated, max_tx_lt FROM updated;
--
--         -- Exit if no rows were updated
--         IF rows_updated = 0 THEN
--             RAISE NOTICE 'No more rows to update: exiting';
--             EXIT;
--         END IF;
--
--         -- Update the last processed tx_lt for the next iteration
--         last_processed_tx_lt := max_tx_lt;
--         iteration_count := iteration_count + 1;
--
--         RAISE NOTICE 'Batch % complete: updated % rows, last_tx_lt = %',
--                      iteration_count, rows_updated, last_processed_tx_lt;
--
--         -- Commit after each batch
--         COMMIT;
--     END LOOP;
--
--     RAISE NOTICE 'Batch update process completed. Total iterations: %', iteration_count;
-- END;
-- $$;
--
-- -- Example usage:
-- -- CALL batch_update_latest_account_states(10000, 0);
