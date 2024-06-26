// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: copyfrom.go

package queries

import (
	"context"
)

// iteratorForCreateBet implements pgx.CopyFromSource.
type iteratorForCreateBet struct {
	rows                 []CreateBetParams
	skippedFirstNextCall bool
}

func (r *iteratorForCreateBet) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForCreateBet) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].UserID,
		r.rows[0].PoolID,
		r.rows[0].Amount,
	}, nil
}

func (r iteratorForCreateBet) Err() error {
	return nil
}

func (q *Queries) CreateBet(ctx context.Context, arg []CreateBetParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"bets"}, []string{"user_id", "pool_id", "amount"}, &iteratorForCreateBet{rows: arg})
}
