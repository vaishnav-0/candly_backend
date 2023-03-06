package handler


import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-redis/redis/v9"
)


type Handlers struct{
	db *pgxpool.Pool 
	rd *redis.Client
}

func NewHandler(db *pgxpool.Pool, rd *redis.Client) *Handlers{
	return &Handlers{
		db:db,
		rd: rd,
	}
}

