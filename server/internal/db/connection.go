package db

import (
	"context"
	"net/url"
	"github.com/jackc/pgx/v5/pgxpool"

)

func Open(ctx context.Context,host, username, password, database string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, "postgresql://"+username+":"+url.QueryEscape(password)+"@"+host+"/"+database)
}
