package market

import "time"

type intervalDet struct {
	Symbol   string
	Interval int64
	Duration time.Duration
}

type interval struct {
	oneMin  intervalDet
	fiveMin intervalDet
}

var Intervals interval = interval{
	oneMin: intervalDet{
		Symbol:   "1m",
		Interval: 60000,
		Duration: time.Minute,
	},
	fiveMin: intervalDet{
		Symbol:   "5m",
		Interval: 300000,
		Duration: time.Minute * 5,
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
	OpenTime int64
	CloseTime int64
}

var PoolTypes = []PoolInfo{
	{"BTUSDT1m", "BTCUSDT", Intervals.oneMin},
	{"BTUSDT5m", "BTCUSDT", Intervals.fiveMin},
}
