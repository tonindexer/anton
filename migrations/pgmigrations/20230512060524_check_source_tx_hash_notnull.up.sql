SET statement_timeout = 0;

ALTER TABLE messages
    ADD CONSTRAINT messages_source_tx_hash_notnull
        CHECK (NOT (source_tx_hash IS NULL AND src_address != decode('11ff0000000000000000000000000000000000000000000000000000000000000000', 'hex')));