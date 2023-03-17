package handler

import (
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)


type Handlers struct{
	db *pgxpool.Pool 
	rd *redis.Client
	log *zerolog.Logger
}

func NewHandler(db *pgxpool.Pool, rd *redis.Client, log *zerolog.Logger) *Handlers{
	return &Handlers{
		db:db,
		rd: rd,
		log: log,
	}
}

