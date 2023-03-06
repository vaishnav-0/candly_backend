package market

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

var mu sync.Mutex
var started uint32

func StartFetchAndStore(store *redis.Client, log *zerolog.Logger) error {

	if atomic.LoadUint32(&started) == 1 {
		return errors.New("already initialized")
	}
	mu.Lock()
	defer mu.Unlock()

	if started == 0 {
		for _, v := range PoolTypes {
			go updateDataPeriodic(v, store, log)
		}
	}

	return nil

}

func updateDataPeriodic(pool PoolInfo, store *redis.Client, log *zerolog.Logger) {

	nextPool, err := PredictNextData(pool.Symbol, pool.Interval.symbol)
	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Type)
		return
	}
	ctx := context.Background()
	guid := xid.New()
	_, err = store.HSet(ctx, pool.Type, "Id", guid.String(), "OpenTime", nextPool.OpenTime, "CloseTime", nextPool.CloseTime).Result()
	store.Expire(ctx, pool.Type, pool.Interval.duration)
	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Type)
	}

	time.AfterFunc(pool.Interval.duration, func() { updateDataPeriodic(pool, store, log) })

}

func GetUpcomingCandle(id string, store *redis.Client) (*CandlestickData, error) {
	ctx := context.Background()
	res, err := store.HMGet(ctx, id, "OpenTime", "CloseTime").Result()
	if err != nil {
		return nil, err
	}

	openTime, ok := res[0].(int64)
	if !ok {
		return nil, errors.New("type not valid")
	}
	closeTime, ok := res[1].(int64)
	if !ok {
		return nil, errors.New("type not valid")
	}

	return &CandlestickData{
		OpenTime:  openTime,
		CloseTime: closeTime,
	}, nil
}
