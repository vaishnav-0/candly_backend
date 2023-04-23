package betting

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"candly/internal/db/queries"
	"candly/internal/market"
	store "candly/internal/memstore"

	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

const BetStatPrefix = "stat:"

func OnUpdate(store *redis.Client, db *pgxpool.Pool, log *zerolog.Logger) (c chan market.UpdatePoolData) {
	ch := make(chan market.UpdatePoolData)

	go func() {
		for poolData := range c {
			err := CreatePool(store, poolData.NewPool.Id, poolData.NewPool.PoolInfo.Interval.Duration)
			if err != nil {
				log.Err(err).Msg("failed to create pool")
			}
			createPoolDB(db, &poolData.PrevPool, log)
		}

	}()

	return ch
}

// type BetData struct {
// 	Id    string
// 	Total int64
// 	Green int64
// 	Red int64
// }

func CreatePool(store *redis.Client, id string, expiry time.Duration) error {
	ctx := context.Background()
	_, err := store.HSet(ctx, id, "id", id).Result()
	if err != nil {
		return fmt.Errorf("failed to create pool. %w", err)
	}
	store.Expire(ctx, id, expiry)
	return nil
}

type BetData struct {
	Total string `json:"stat:total"`
	Red   string `json:"stat:red"`
	Green string `json:"stat:green"`
	User  string `json:"user1"`
}

func Bet(store *redis.Client, id string, user string, amount int64) error {
	ctx := context.Background()
	pipe := store.Pipeline()
	exist := pipe.Exists(ctx, id)
	betAmt := pipe.HGet(ctx, id, user)

	_, err := pipe.Exec(ctx)

	if err != nil && err != redis.Nil {
		return err
	}

	var bet int64 = 0
	if betAmt.Val() == "" {
		bet = 0
	} else {
		bet, err = strconv.ParseInt(betAmt.Val(), 10, 64)
		if err != nil {
			return err
		}
	}

	if exist.Val() != 1 {
		return PoolNotFoundError
	}

	_ = pipe.HSet(ctx, id, user, amount)
	diff := int64(math.Abs(float64(amount))) - int64(math.Abs(float64(bet)))
	_ = pipe.HIncrBy(ctx, id, BetStatPrefix+"total", diff)
	redKey := BetStatPrefix + "red"
	greenKey := BetStatPrefix + "green"
	if bet < 0 {
		if amount < 0 {
			_ = pipe.HIncrBy(ctx, id, redKey, diff)
		} else {
			_ = pipe.HIncrBy(ctx, id, redKey, -bet)
			_ = pipe.HIncrBy(ctx, id, greenKey, amount)
		}
	} else {
		if amount >= 0 {
			_ = pipe.HIncrBy(ctx, id, greenKey, diff)
		} else {
			_ = pipe.HIncrBy(ctx, id, redKey, int64(math.Abs(float64(amount))))
			_ = pipe.HIncrBy(ctx, id, greenKey, -bet)
		}
	}

	_, err = pipe.Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func GetBets(rd *redis.Client, id string) (map[string]string, error) {
	return store.GetHash(rd, id)
}

func GetPools(store *redis.Client) ([]map[string]string, error) {
	ret := make([]map[string]string, 0, 5)
	for _, poolType := range market.PoolTypes {
		ctx := context.Background()
		res, err := store.HGetAll(ctx, poolType.Type).Result()

		if len(res) == 0 {
			continue
		}

		if err != nil {
			return nil, err
		}
		ret = append(ret, res)
	}
	return ret, nil
}

func createPoolDB(db *pgxpool.Pool, pool *market.Pool, log *zerolog.Logger) {
	q := queries.New(db)
	ctx := context.Background()
	openTime := pgtype.Int8{}
	openTime.Scan(pool.OpenTime)

	closeTime := pgtype.Int8{}
	closeTime.Scan(pool.CloseTime)

	err := q.CreatePool(ctx, queries.CreatePoolParams{
		ID:        pool.Id,
		OpenTime:  openTime,
		CloseTime: closeTime,
		Type:      queries.PoolType(pool.PoolInfo.Interval.Symbol),
	})

	if err != nil {
		log.Err(err).Msg("failed to insert pool")
	}
}
