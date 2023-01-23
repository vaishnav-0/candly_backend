package market

type intervalDet struct{
	symbol string
	interval int64
}

type interval struct{
	oneMin intervalDet
	fiveMin intervalDet
}

var Intervals interval = interval{
	oneMin: intervalDet{
		symbol:"1m",
		interval: 60000,
	},
	fiveMin: intervalDet{
		symbol:"5m",
		interval: 300000,
	},
}

type Pool struct{
	Id string
	Symbol string
	Interval intervalDet
}

var Pools = []Pool{
	{"BTUSDT1m", "BTCUSDT", Intervals.oneMin},
	{"BTUSDT5m", "BTCUSDT", Intervals.fiveMin},
}