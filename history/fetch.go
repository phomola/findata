package history

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type (
	quote struct {
		Volume        []float64 `json:"volume"`
		Open          []float64 `json:"open"`
		Close         []float64 `json:"close"`
		Low           []float64 `json:"low"`
		High          []float64 `json:"high"`
		AdjustedClose []float64 `json:"adjclose"`
	}

	indicators struct {
		Quotes []quote `json:"quote"`
	}

	// Meta contains data about financial data such as currency, exchange etc.
	Meta struct {
		Currency             string `json:"currency"`
		Symbol               string `json:"symbol"`
		ExchangeName         string `json:"exchangeName"`
		FullExchangeName     string `json:"fullExchangeName"`
		Timezone             string `json:"timezone"`
		ExchangeTimezoneName string `json:"exchangeTimezoneName"`
		LongName             string `json:"longName"`
		ShortName            string `json:"shortName"`
	}

	result struct {
		Timestamps []int64    `json:"timestamp"`
		Indicators indicators `json:"indicators"`
		Meta       Meta       `json:"meta"`
	}

	chart struct {
		Results []result `json:"result"`
	}

	response struct {
		Chart chart `json:"chart"`
	}

	// Candle contains candle data.
	Candle struct {
		Timestamp     time.Time
		Volume        float64
		Open          float64
		Close         float64
		Low           float64
		High          float64
		AdjustedClose float64
	}
)

// Interval represents a time interval.
type Interval int

const (
	// Interval1D is the 1-day interval.
	Interval1D Interval = iota
)

func (i Interval) String() string {
	switch i {
	case Interval1D:
		return "1d"
	}
	return ""
}

// Fetch fetches financial data.
func Fetch(symbol string, from, to time.Time, interval Interval) ([]Candle, *Meta, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s", symbol, from.Unix(), to.Unix(), interval)
	cl := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := cl.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected error code: %s", resp.Status)
	}
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, nil, err
	}
	timestamps := r.Chart.Results[0].Timestamps
	volume := r.Chart.Results[0].Indicators.Quotes[0].Volume
	open := r.Chart.Results[0].Indicators.Quotes[0].Open
	close := r.Chart.Results[0].Indicators.Quotes[0].Close
	low := r.Chart.Results[0].Indicators.Quotes[0].Low
	high := r.Chart.Results[0].Indicators.Quotes[0].High
	adjclose := r.Chart.Results[0].Indicators.Quotes[0].AdjustedClose
	meta := r.Chart.Results[0].Meta
	loc, err := time.LoadLocation(meta.ExchangeTimezoneName)
	if err != nil {
		return nil, nil, err
	}
	candles := make([]Candle, 0, len(timestamps))
	for i, t := range timestamps {
		candles = append(candles, Candle{
			Timestamp:     time.Unix(t, 0).In(loc),
			Volume:        volume[i],
			Open:          open[i],
			Close:         close[i],
			Low:           low[i],
			High:          high[i],
			AdjustedClose: adjclose[i],
		})
	}
	return candles, &meta, nil
}
