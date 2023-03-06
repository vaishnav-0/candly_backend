package market

import "time"

type intervalDet struct {
	symbol   string
	interval int64
	duration time.Duration
}

type interval struct {
	oneMin  intervalDet
	fiveMin intervalDet
}

var Intervals interval = interval{
	oneMin: intervalDet{
		symbol:   "1m",
		interval: 60000,
		duration: time.Minute,
	},
	fiveMin: intervalDet{
		symbol:   "5m",
		interval: 300000,
		duration: time.Minute * 5,
	},
}

type PoolInfo struct {
	Type     string
	Symbol   string
	Interval intervalDet
}

type Pool struct {
	PoolInfo
	Id string
}

var PoolTypes = []PoolInfo{
	{"BTUSDT1m", "BTCUSDT", Intervals.oneMin},
	{"BTUSDT5m", "BTCUSDT", Intervals.fiveMin},
}
