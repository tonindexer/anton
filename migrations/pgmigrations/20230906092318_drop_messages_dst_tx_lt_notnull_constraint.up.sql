SET statement_timeout = 0;

--bun:split

ALTER TABLE messages DROP CONSTRAINT messages_tx_lt_notnull;

--bun:split

ALTER TABLE messages ADD CONSTRAINT messages_tx_lt_notnull CHECK ((
    (
        (type = 'EXTERNAL_OUT') AND (src_address IS NOT NULL) AND (src_tx_lt IS NOT NULL) AND (dst_address IS NULL) AND (dst_tx_lt IS NULL)
    ) OR (
        (type = 'EXTERNAL_IN') AND (src_address IS NULL) AND (src_tx_lt IS NULL) AND (dst_address IS NOT NULL) AND (dst_tx_lt IS NOT NULL)
    ) OR (
        (type = 'INTERNAL') AND (src_workchain != -1 OR dst_workchain != -1) AND (src_tx_lt IS NOT NULL)
    ) OR (
        (type = 'INTERNAL') AND src_workchain = -1 AND dst_workchain = -1
    )
));
