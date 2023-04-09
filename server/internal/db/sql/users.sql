-- name: InsertUser :exec
INSERT INTO users(name, phone, email)
VALUES ($1, $2, $3);

-- name: GetUser :one
SELECT * FROM users WHERE id=$1;

-- name: GetUserFromPhone :one
SELECT * FROM users WHERE phone=$1;