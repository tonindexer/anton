SET statement_timeout = 0;

ALTER TABLE messages
    DROP CONSTRAINT messages_source_tx_hash_notnull;