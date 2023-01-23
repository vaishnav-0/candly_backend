package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"net/url"
)

func Open(host, username, password, database string) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), "postgresql://"+username+":"+url.QueryEscape(password)+"@"+host+"/"+database)
}
