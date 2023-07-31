package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

//needs to be int64 to work
func PgInt8(val int64)(casted pgtype.Int8){
	casted.Scan(val)
	return
}

func PgInt4(val int64 )(casted pgtype.Int4){
	casted.Scan(val)
	return
}

func PgText(val string)(casted pgtype.Text){
	casted.Scan(val)
	return
}