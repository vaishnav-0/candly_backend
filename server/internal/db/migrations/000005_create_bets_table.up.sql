CREATE TABLE IF NOT EXISTS bets(
    id VARCHAR(22) PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ,
	pool_id VARCHAR(25) REFERENCES pools(id),
	amount INTEGER, 
	won BOOLEAN 
);


CREATE INDEX index_bets_unwon ON bets(id) WHERE won IS NULL;