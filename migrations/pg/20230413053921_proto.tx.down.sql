SET statement_timeout = 0;

--bun:split
DROP TYPE account_status;

DROP TABLE latest_account_states;

DROP TABLE account_states;

DROP TABLE account_data;

--bun:split

DROP TYPE message_type;

DROP TABLE messages;

DROP TABLE message_payloads;

DROP TABLE transactions;

--bun:split

DROP TABLE block_info;

--bun:split

DROP TABLE contract_interfaces;

DROP TABLE contract_operations;
