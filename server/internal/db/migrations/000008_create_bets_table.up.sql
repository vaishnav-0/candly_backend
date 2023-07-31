CREATE TABLE IF NOT EXISTS bets(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY ,
    user_id VARCHAR(100) REFERENCES users(name) ,
	pool_id VARCHAR(25) REFERENCES pools(id),
	amount INTEGER, 
	won BOOLEAN 
);


CREATE INDEX index_bets_unwon ON bets(id) WHERE won IS NULL;