SET statement_timeout = 0;

--bun:split

ALTER TABLE account_states ADD COLUMN fake boolean NOT NULL DEFAULT false;