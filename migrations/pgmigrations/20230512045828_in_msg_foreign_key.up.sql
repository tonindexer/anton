ALTER TABLE ONLY transactions
    ADD CONSTRAINT transactions_in_msg_hash_fkey FOREIGN KEY (in_msg_hash) REFERENCES messages(hash);
