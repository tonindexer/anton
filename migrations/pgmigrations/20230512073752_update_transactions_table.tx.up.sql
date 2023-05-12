SET statement_timeout = 0;

ALTER TABLE transactions DROP COLUMN state_update;

--bun:split

ALTER TABLE transactions ADD COLUMN result_code integer NOT NULL DEFAULT 0;

--bun:split

ALTER TABLE transactions ALTER COLUMN description SET NOT NULL;
