-- name: CreatePool :exec
INSERT INTO pools (
  id, open_time, close_time, type
) VALUES (
  $1, $2, $3, $4
);

-- name: DeletePool :exec
DELETE FROM pools where id=$1;

-- name: UpdatePool :exec
UPDATE pools SET open_price=$2, close_price=$3
WHERE id=$1;

-- name: CreateBet :copyfrom
INSERT INTO bets (
  user_id, pool_id, amount
) VALUES (
  $1, $2, $3
);

-- name: CreateBetStat :exec
INSERT INTO bet_stats (
  pool_id, total, red, green, total_bets, red_bets, green_bets
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);