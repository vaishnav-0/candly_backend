CREATE TABLE IF NOT EXISTS pools(
    id VARCHAR(25) PRIMARY KEY,
    open_time BIGINT,
	close_time BIGINT,
	open_price NUMERIC(20, 10), 
	close_price NUMERIC(20, 10),
    completed BOOLEAN DEFAULT false
)