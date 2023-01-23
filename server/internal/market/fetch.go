package market

import (
	er "candly/internal/errors"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/buger/jsonparser"
)

type CandlestickData struct {
	OpenTime                 int64
	OpenPrice                float64
	HighPrice                string
	LowPrice                 string
	ClosePrice               float64
	Volume                   string
	CloseTime           	 int64
	QuoteAssetVolume         string
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
}

func GetLatestCandleData(symbol string, interval string) (*CandlestickData, error) {
	return GetCandleData(symbol, interval, "1", "", "")
}

func GetCandleData(symbol string, interval string, limit string, startTime string, endTime string) (*CandlestickData, error) {
	req, err := http.NewRequest("GET", "https://api3.binance.com/api/v3/klines", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required. %w", er.ErrorFatal)
	}
	if interval == "" {
		return nil, fmt.Errorf("interval is required. %w", er.ErrorFatal)
	}
	q.Add("symbol", symbol)
	q.Add("interval", interval)

	if limit != "" {
		q.Add("limit", limit)
	}
	if endTime != "" {
		q.Add("endTime", endTime)
	}
	if startTime != "" {
		q.Add("startTime", startTime)
	}

	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}


	data, jsonErr := UnmarshalJSON(body)
	if jsonErr != nil {
		return nil, fmt.Errorf("%w; %s", err, string(body))
	}

	return data, nil

}

func PredictNextData(symbol string, interval string) (*CandlestickData, error) {
	currentData, err := GetLatestCandleData(symbol, interval)
	if err != nil {
		return nil, err
	}

	var timeIncr int64
	if(interval == Intervals.oneMin.symbol){
		timeIncr = Intervals.oneMin.interval
	}else if(interval == Intervals.fiveMin.symbol){
		timeIncr = Intervals.fiveMin.interval
	}else{
		return nil, errors.New("time interval not found")
	}


	return &CandlestickData{
		OpenTime: currentData.CloseTime + 1,
		CloseTime: currentData.CloseTime + timeIncr,
	}, nil
}

func UnmarshalJSON(bs []byte) (*CandlestickData, error) {
	var err error
	var data CandlestickData
	paths := [][]string{
		{"[0]", "[0]"},
		{"[0]", "[1]"},
		{"[0]", "[2]"},
		{"[0]", "[3]"},
		{"[0]", "[4]"},
		{"[0]", "[5]"},
		{"[0]", "[6]"},
		{"[0]", "[7]"},
		{"[0]", "[8]"},
		{"[0]", "[9]"},
		{"[0]", "[10]"},
	}
	jsonparser.EachKey(bs, func(idx int, value []byte, vt jsonparser.ValueType, er error) {
		if er != nil {
			err = er
			return
		}
		switch idx {
		case 0:
			data.OpenTime, err = jsonparser.ParseInt(value)
		case 1:
			data.OpenPrice, err = strconv.ParseFloat(string(value), 32)

		case 2:
			data.HighPrice = string(value)
		case 3:
			data.LowPrice = string(value)

		case 4:
			data.ClosePrice, err = strconv.ParseFloat(string(value), 32)
		case 5:

			data.Volume = string(value)
		case 6:
			data.CloseTime, err = jsonparser.ParseInt(value)
		case 7:
			data.QuoteAssetVolume = string(value)
		case 8:
			data.NumberOfTrades, err = jsonparser.ParseInt(value)
		case 9:
			data.TakerBuyBaseAssetVolume = string(value)
		case 10:
			data.TakerBuyQuoteAssetVolume = string(value)
		}

	}, paths...)
	if err != nil {
		return nil, err
	}
	return &data, err
}
