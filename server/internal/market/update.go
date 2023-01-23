package market

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v9"
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
		for _, v := range Pools {
			go updateDataPeriodic(v, store, log)
		}
	}

	return nil

}

func updateDataPeriodic(pool Pool, store *redis.Client, log *zerolog.Logger) {

	nextPool, err := PredictNextData(pool.Symbol, pool.Interval.symbol)
	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Id)
		return
	}
	ctx := context.Background()

	_, err = store.HSet(ctx, pool.Id, "OpenTime", nextPool.OpenTime, "CloseTime", nextPool.CloseTime).Result()

	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Id)
	}

	time.AfterFunc(time.Duration(time.Now().UnixNano())-time.Duration(nextPool.OpenTime*1000000), func() { updateDataPeriodic(pool, store, log) })

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
