SET statement_timeout = 0;

ALTER TABLE ONLY transactions
    DROP CONSTRAINT transactions_in_msg_hash_fkey;
