BEGIN;

    ALTER TABLE latest_account_states
        ADD COLUMN created_lt bigint;

COMMIT;


-- --bun:split
-- CREATE OR REPLACE PROCEDURE batch_fill_account_states_created_lt(
--     batch_size INT DEFAULT 10000,
--     start_from_address BYTEA DEFAULT NULL
-- )
-- LANGUAGE plpgsql
-- AS $$
-- DECLARE
--     last_processed_address BYTEA := start_from_address;
--     rows_updated INT;
--     iteration_count INT := 0;
--     max_address BYTEA;
-- BEGIN
--     LOOP
--         -- Directly query account_states for minimum last_tx_lt per address
--         WITH min_tx_lt AS (
--             SELECT address, MIN(last_tx_lt) as min_lt
--             FROM account_states
--             WHERE address > last_processed_address
--             GROUP BY address
--             ORDER BY address
--             LIMIT batch_size
--         ),
--         updated AS (
--             UPDATE latest_account_states las
--             SET created_lt = m.min_lt
--             FROM min_tx_lt m
--             WHERE las.address = m.address
--               AND las.created_lt IS NULL
--             RETURNING las.address
--         ),
--         last_address AS (
--             SELECT address FROM updated ORDER BY address DESC LIMIT 1
--         ),
--         address_count AS (
--             SELECT COUNT(*) AS count FROM updated
--         )
--         SELECT address_count.count, last_address.address INTO rows_updated, max_address FROM address_count, last_address;
--
--         -- Exit if no rows were updated
--         IF COALESCE(rows_updated, 0) = 0 THEN
--             RAISE NOTICE 'No more rows to update: exiting';
--             EXIT;
--         END IF;
--
--         -- Update the last processed address for the next iteration
--         last_processed_address := max_address;
--         iteration_count := iteration_count + 1;
--
--         RAISE NOTICE 'Batch % complete: updated % rows, last address = %',
--                      iteration_count, rows_updated, encode(last_processed_address, 'hex');
--
--         -- Commit after each batch
--         COMMIT;
--     END LOOP;
--
--     RAISE NOTICE 'Batch fill process completed. Total iterations: %', iteration_count;
-- END;
-- $$;
--
-- -- Example usage:
-- -- CALL batch_fill_account_states_created_lt(60000, decode('000000000000000000000000000000000000000000000000000000000000000000', 'hex'));
-- -- CALL batch_fill_account_states_created_lt(60000, decode('0074b000e63938eb4547be7a5c3011ec6c5cb2fc80f55539b8124c5e4e5851818a', 'hex'));
