package handler

import (
	"candly/internal/auth"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)


type Handlers struct{
	db *pgxpool.Pool 
	rd *redis.Client
	log *zerolog.Logger
	auth *auth.Auth
}

func NewHandler(db *pgxpool.Pool, rd *redis.Client, log *zerolog.Logger, auth *auth.Auth) *Handlers{
	return &Handlers{
		db:db,
		rd: rd,
		log: log,
		auth: auth,
	}
}

