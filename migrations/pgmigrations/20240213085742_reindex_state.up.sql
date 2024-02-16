SET statement_timeout = 0;

BEGIN;
    CREATE SEQUENCE rescan_tasks_id_seq START WITH 1;

    CREATE TABLE rescan_tasks (
        id integer NOT NULL DEFAULT nextval('rescan_tasks_id_seq'),
        finished bool NOT NULL,
        start_from_masterchain_seq_no integer NOT NULL,
        accounts_last_masterchain_seq_no integer NOT NULL,
        accounts_rescan_done boolean NOT NULL,
        messages_last_masterchain_seq_no integer NOT NULL,
        messages_rescan_done boolean NOT NULL,
        CONSTRAINT rescan_tasks_pkey PRIMARY KEY (id)
    );

    ALTER SEQUENCE rescan_tasks_id_seq OWNED BY rescan_tasks.id;

    CREATE UNIQUE INDEX ON rescan_tasks (finished) WHERE finished = false;
COMMIT;

CREATE INDEX account_states_workchain_shard_block_seq_no_idx ON account_states USING btree (workchain, shard, block_seq_no);
