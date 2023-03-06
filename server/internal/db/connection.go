package db

import (
	"context"
	"net/url"
	"github.com/jackc/pgx/v5/pgxpool"

)

func Open(host, username, password, database string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), "postgresql://"+username+":"+url.QueryEscape(password)+"@"+host+"/"+database)
}
