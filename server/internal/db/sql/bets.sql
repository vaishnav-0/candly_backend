-- name: CreatePool :exec
INSERT INTO pools (
  id, open_time, close_time, type
) VALUES (
  $1, $2, $3, $4
);

-- name: UpdatePool :exec
UPDATE pools SET open_price=$2, close_price=$3
WHERE id=$1;
