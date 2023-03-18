package market

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"candly/internal/memstore"

	"github.com/go-redis/redis/v9"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

type UpdatePoolData struct {
	NewPool  Pool
	PrevPool Pool
}

var mu sync.Mutex
var started uint32

func StartFetchAndStore(store *redis.Client, log *zerolog.Logger, ch chan<- UpdatePoolData) error {

	if atomic.LoadUint32(&started) == 1 {
		return errors.New("already initialized")
	}
	mu.Lock()
	defer mu.Unlock()

	if started == 0 {
		for _, v := range PoolTypes {
			go updateDataPeriodic(v, store, log, ch)
		}
	}

	return nil

}

func updateDataPeriodic(pool PoolInfo, store *redis.Client, log *zerolog.Logger, ch chan<- UpdatePoolData) {

	nextPool, err := PredictNextData(pool.Symbol, pool.Interval.Symbol)
	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Type)
		return
	}

	time.Sleep(time.Until(time.UnixMilli(nextPool.OpenTime)))
	go updateData(pool, store, log, ch)

	ticker := time.NewTicker(pool.Interval.Duration)

	for i := range ticker.C {
		fmt.Println(i)
		go updateData(pool, store, log, ch)
	}

}

func updateData(pool PoolInfo, store *redis.Client, log *zerolog.Logger, ch chan<- UpdatePoolData) {

	nextPool, err := PredictNextData(pool.Symbol, pool.Interval.Symbol)
	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Type)
		return
	}

	prevPoolData, err := memstore.GetHash(store, pool.Type)

	if err != nil {
		log.Err(err).Msg("Failed to get pool previous pool id: " + pool.Type)
		return
	}

	prevClose, _ := strconv.ParseInt(prevPoolData["closeTime"], 10, 64)
	prevOpen, _ := strconv.ParseInt(prevPoolData["openTime"], 10, 64)

	prevPool := Pool{
		Id:        prevPoolData["id"],
		PoolInfo:  pool,
		OpenTime:  prevOpen,
		CloseTime: prevClose,
	}

	guid := xid.New()
	ctx := context.Background()
	_, err = store.HSet(ctx, pool.Type, "id", guid.String(), "openTime", nextPool.OpenTime, "closeTime", nextPool.CloseTime).Result()
	store.Expire(ctx, pool.Type, pool.Interval.Duration+time.Second*10)

	if err != nil {
		log.Err(err).Msg("Failed to get next pool " + pool.Type)
		return
	}

	if prevPool.Id == ""{
		return
	}
	ch <- UpdatePoolData{
		NewPool: Pool{
			Id:        guid.String(),
			OpenTime:  nextPool.OpenTime,
			CloseTime: nextPool.CloseTime,
			PoolInfo:  pool,
		},
		PrevPool: prevPool,
	}
}

func GetPoolId(store *redis.Client, pool_type string) (string, error) {
	return store.HGet(context.Background(), pool_type, "id").Result()
}

// func GetUpcomingCandle(id string, store *redis.Client) (*CandlestickData, error) {
// 	ctx := context.Background()
// 	res, err := store.HMGet(ctx, id, "OpenTime", "CloseTime").Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	openTime, ok := res[0].(int64)
// 	if !ok {
// 		return nil, errors.New("type not valid")
// 	}
// 	closeTime, ok := res[1].(int64)
// 	if !ok {
// 		return nil, errors.New("type not valid")
// 	}

// 	return &CandlestickData{
// 		OpenTime:  openTime,
// 		CloseTime: closeTime,
// 	}, nil
// }
