SET statement_timeout = 0;

BEGIN;
    CREATE TYPE rescan_task_type AS ENUM (
        'add_interface',
        'upd_interface',
        'del_interface',
        'add_get_method',
        'del_get_method',
        'upd_get_method',
        'upd_operation',
        'del_operation'
    );

    CREATE SEQUENCE rescan_tasks_id_seq START WITH 1;

    CREATE TABLE rescan_tasks (
        id integer NOT NULL DEFAULT nextval('rescan_tasks_id_seq'),
        finished bool NOT NULL,
        type rescan_task_type NOT NULL,

        contract_name text NOT NULL,

        changed_get_method text NOT NULL,

        message_type message_type NOT NULL,
        outgoing boolean NOT NULL,
        operation_id integer NOT NULL,

        last_address bytea NOT NULL,
        last_tx_lt bigint NOT NULL,

        CONSTRAINT rescan_tasks_pkey PRIMARY KEY (id)
    );

    ALTER SEQUENCE rescan_tasks_id_seq OWNED BY rescan_tasks.id;
COMMIT;
