CREATE TABLE bet_stats(
    pool_id VARCHAR(25) PRIMARY KEY REFERENCES pools(id),
    total INTEGER,
    red INTEGER,
    green INTEGER,
    total_bets INTEGER,
    red_bets INTEGER,
    green_bets INTEGER
);