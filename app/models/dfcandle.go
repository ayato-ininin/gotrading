package models

import "time"

//df = データフレーム
//これはCandleの中のopen、あるいは、closeの値だけスライスで返すといったことをする。

type DataFlameCandle struct{
	ProductCode string `json:"product_code"`
	Duration time.Duration `json:"duration"`
	Candles []Candle `json:"candles"`
}

/*
Candleのデータから、timeのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Times() []time.Time{
	s := make([]time.Time, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.Time
	}
	return s
}

/*
Candleのデータから、openのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Opens() []float64{
	s := make([]float64, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.Open
	}
	return s
}

/*
Candleのデータから、closeのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Closes() []float64{
	s := make([]float64, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.Close
	}
	return s
}

/*
Candleのデータから、highのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Highs() []float64{
	s := make([]float64, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.High
	}
	return s
}

/*
Candleのデータから、lowのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Lows() []float64{
	s := make([]float64, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.Low
	}
	return s
}

/*
Candleのデータから、volumeのデータのみスライスで返す。
*/
func (df *DataFlameCandle) Volume() []float64{
	s := make([]float64, len(df.Candles))
	for i,candle := range df.Candles{
		s[i] = candle.Volume
	}
	return s
}
