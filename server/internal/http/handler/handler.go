package handler


import (
	"github.com/jackc/pgx/v5"
	"github.com/go-redis/redis/v9"
)


type Handlers struct{
	db *pgx.Conn 
	rd *redis.Client
}

func NewHandler(db *pgx.Conn, rd *redis.Client) *Handlers{
	return &Handlers{
		db:db,
		rd: rd,
	}
}

