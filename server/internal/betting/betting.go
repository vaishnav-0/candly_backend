package betting

import (
	"fmt"
	"math"
	"strconv"

	"candly/internal/db/queries"
	"candly/internal/market"
	store "candly/internal/memstore"

	dbPkg "candly/internal/db"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonhoo/go-events"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"
)

const BetStatPrefix = "stat:"
const RedKey = BetStatPrefix + "red"
const GreenKey = BetStatPrefix + "green"
const TotalKey = BetStatPrefix + "total"

func OnUpdate(store *redis.Client, db *pgxpool.Pool, log *zerolog.Logger) (c chan market.UpdatePoolData) {
	c = make(chan market.UpdatePoolData)

	go func() {
		for poolData := range c {

			err := CreatePool(store, poolData.NewPool.Id)
			if err != nil {
				log.Err(err).Msg("failed to create pool")
			}
			//order matters due to foriegn key
			createPoolDB(db, &poolData.PrevPool, log)
			go saveBets(store, db, poolData.PrevPool.Id, log)

		}

	}()

	return
}

func saveBets(store *redis.Client, db *pgxpool.Pool, id string, log *zerolog.Logger) {
	ctx := context.Background()

	size, err := store.HLen(ctx, id).Result()
	fmt.Println(size)
	if err != nil {
		log.Err(err).Msg("failed to store bets")
	}

	// no key. Possible on the first run
	if size == 0 {
		return
	}

	q := queries.New(db)

	//no bets
	if size == 4 {
		store.Del(ctx, id)
		q.DeletePool(ctx, id)
		return
	}

	pipe := store.Pipeline()

	totalCmd := pipe.HGet(ctx, id, TotalKey)
	redCmd := pipe.HGet(ctx, id, RedKey)
	greenCmd := pipe.HGet(ctx, id, GreenKey)

	pipe.HDel(ctx, id, TotalKey, RedKey, GreenKey, "id")

	betsCmd := pipe.HGetAll(ctx, id)

	_ , err = pipe.Exec(ctx)


	if err != nil {
		log.Err(err).Msg("failed to store bets(falure in executing redis commands)")
	}

	total, _ := strconv.ParseInt(totalCmd.Val(), 10, 64)
	red, _ := strconv.ParseInt(redCmd.Val(), 10, 64)
	green, _ := strconv.ParseInt(greenCmd.Val(), 10, 64)

	len := int64(size - 4) // others are bets

	bets := betsCmd.Val()


	allBets := make([]queries.CreateBetParams, 0, len)

	var totalRed int64 = 0
	var totalGreen int64 = 0


	fmt.Println(allBets)

	for k, v := range bets {
		fmt.Println(k)
		userId:= k
		val, _ := strconv.ParseInt(v, 10, 64)

		if val < 0 {
			totalRed++
		} else {
			totalGreen++
		}

		allBets = append(allBets, queries.CreateBetParams{
			PoolID: dbPkg.PgText(id),
			UserID: dbPkg.PgText(userId),
			Amount: dbPkg.PgInt4(val),
		})
	}

	fmt.Println(allBets)


	_, err = q.CreateBet(ctx, allBets)

	if err != nil {
		log.Err(err).Msg("failed to store bets to db")
		return
	}

	err = q.CreateBetStat(ctx, queries.CreateBetStatParams{
		PoolID:    id,
		Red:       dbPkg.PgInt4(red),
		Green:     dbPkg.PgInt4(green),
		Total:     dbPkg.PgInt4(total),
		TotalBets: dbPkg.PgInt4(len),
		RedBets:   dbPkg.PgInt4(totalRed),
		GreenBets: dbPkg.PgInt4(totalGreen),
	})

	if err != nil {
		log.Err(err).Msg("failed to store bet stats to db")
		return
	}

	store.Del(ctx, id)
}

// type BetData struct {
// 	Id    string
// 	Total int64
// 	Green int64
// 	Red int64
// }

func CreatePool(store *redis.Client, id string) error {
	ctx := context.Background()
	pipe := store.Pipeline()
	pipe.HSet(ctx, id, "id", id)
	pipe.HSet(ctx, id, GreenKey, 0)
	pipe.HSet(ctx, id, TotalKey, 0)
	pipe.HSet(ctx, id, RedKey, 0)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create pool. %w", err)
	}

	// store.Expire(ctx, id, expiry)
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

	SendNewBetEvent(id, user, amount, 0, 0, 0)

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

type EventType string

const (
	NewBetEvent EventType = "NewBet"
)

type NewBet struct {
	User   string
	Amount int64
	Total  int
	Green  int
	Red    int
	Type   EventType
}

func SendNewBetEvent(id string, user string, amount int64, total int, green int, red int) {
	sendEvent(id, NewBet{
		Type:   NewBetEvent,
		User:   user,
		Amount: amount,
		Total:  total,
		Red:    red,
		Green:  green,
	})
}

func sendEvent(event string, data interface{}) {
	events.Announce(events.Event{Tag: event, Data: data})
}
