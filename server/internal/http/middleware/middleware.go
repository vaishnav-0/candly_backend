package middleware

import (
	"candly/internal/auth"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)


type Middlewares struct{
	db *pgxpool.Pool 
	rd *redis.Client
	log *zerolog.Logger
	auth *auth.Auth
}

func NewMiddleware(db *pgxpool.Pool, rd *redis.Client, log *zerolog.Logger, auth *auth.Auth) *Middlewares{
	return &Middlewares{
		db:db,
		rd: rd,
		log: log,
		auth: auth,
	}
}


