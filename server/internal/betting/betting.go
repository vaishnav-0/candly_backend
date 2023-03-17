package betting

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"candly/internal/market"
	store "candly/internal/memstore"

	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

func OnUpdateFn(store *redis.Client, log *zerolog.Logger) market.OnUpdate {
	return func(id string, openTime int64, closeTime int64, poolDuration time.Duration) {
		err := CreatePool(store, id, poolDuration)
		if err != nil {
			log.Err(err).Msg("failed to create pool")
		}
	}
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
			return fmt.Errorf("errorrrrrr. %w", err)
			return err
		}
	}

	if exist.Val() != 1 {
		return errors.New("pool not found")
	}

	_ = pipe.HSet(ctx, id, user, amount)
	diff := int64(math.Abs(float64(amount))) - int64(math.Abs(float64(bet)))
	_ = pipe.HIncrBy(ctx, id, "total", diff)
	if bet < 0 {
		if amount < 0 {
			_ = pipe.HIncrBy(ctx, id, "red", diff)
		} else {
			_ = pipe.HIncrBy(ctx, id, "red", -bet)
			_ = pipe.HIncrBy(ctx, id, "green", amount)
		}
	} else {
		if amount >= 0 {
			_ = pipe.HIncrBy(ctx, id, "green", diff)
		} else {
			_ = pipe.HIncrBy(ctx, id, "red", int64(math.Abs(float64(amount))))
			_ = pipe.HIncrBy(ctx, id, "green", -bet)
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

		if err != nil {
			return nil, err
		}
		ret = append(ret, res)
	}
	return ret, nil
}
