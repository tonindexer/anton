SET statement_timeout = 0;

CREATE TABLE contract_definitions (
    name text NOT NULL,
    schema jsonb NOT NULL,
    CONSTRAINT contract_definitions_pkey PRIMARY KEY (name)
);
