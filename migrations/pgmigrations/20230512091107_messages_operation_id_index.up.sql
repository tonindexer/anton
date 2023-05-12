SET statement_timeout = 0;

CREATE INDEX messages_operation_id_idx ON messages USING hash (operation_id);