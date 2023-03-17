package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
)

type Config struct {
	Host     string `env:"DBHOST,notEmpty"`
	Username string `env:"DBUSERNAME,notEmpty"`
	Name     string `env:"DBNAME,notEmpty"`
	Password string `env:"DBPASSWORD,notEmpty"`
}

func Open(ctx context.Context, host, username, password, database string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, "postgresql://"+username+":"+url.QueryEscape(password)+"@"+host+"/"+database)
}
