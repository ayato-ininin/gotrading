package models

import (
	"fmt"
	"gotrading/bitflyer"
	"time"
)

// jsonでreturnされうる構造体なので、json用に書く
type Candle struct {
	ProductCode string        `json:"product_code"`
	Duration    time.Duration `json:"duration"`
	Time        time.Time     `json:"time"`
	Open        float64       `json:"open"`
	Close       float64       `json:"close"`
	High        float64       `json:"high"`
	Low         float64       `json:"low"`
	Volume      float64       `json:"volume"`
}

func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open, close, high, low, volume float64) *Candle {
	return &Candle{
		productCode,
		duration,
		timeDate,
		open,
		close,
		high,
		low,
		volume,
	}
}

// table名を取得する
// 1h,1m,1sで共通の処理。
func (c *Candle) TableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

func (c *Candle) Create() error {
	cmd := fmt.Sprintf("INSERT INTO %s (time,open,close,high,low,volume) VALUES (?,?,?,?,?,?)", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Time.Format(time.RFC3339), c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		return err
	}
	return err
}

/*
updateメソッド
1h,1m,1sでかぶってるのがあれば、データを更新する。
timestampがkeyとなる。
*/
func (c *Candle) Save() error {
	cmd := fmt.Sprintf("UPDATE %s SET open = ?,close=?,high = ?,low= ?,volume = ? WHERE time = ?", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return err
}

func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *Candle {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf("SELECT time, open,close,high,low,volume FROM time = ?", tableName)
	row := DbConnection.QueryRow(cmd, dateTime.Format(time.RFC3339)) //一行のみ返す
	var candle Candle
	err := row.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil {
		return nil
	}
	return NewCandle(productCode, duration, candle.Time, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}

/*
リアルタイム通信でtickerが何度も来るから、その度に発火する。
データベースで新たなrowを作成したらtrue
データを作成せず、更新だけならfalse
*/
func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	//まずDBにデータがあるのかを確認。1mなら、1m15sとかもすべて1mとしてtruncateされてくる
	currentCandle := GetCandle(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	//データがない場合
	if currentCandle == nil {
		candle := NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		candle.Create()
		return true
	}
	//あれば更新のために条件分岐
	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	//volumeは合計値
	currentCandle.Volume += ticker.Volume
	currentCandle.Close = price
	currentCandle.Save()
	return false
}

func GetAllCandle(productCode string, duration time.Duration, limit int) (dfCandle *DataFlameCandle, err error) {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time,open,close,high,low,volume FROM %s ORDER BY time DESC LIMIT ?
		) ORDER BY time ASC;`, tableName)
	rows, err := DbConnection.Query(cmd, limit)
	if err != nil {
		return
	}
	defer rows.Close()

	dfCandle = &DataFlameCandle{}
	dfCandle.ProductCode = productCode
	dfCandle.Duration = duration
	for rows.Next() {
		var candle Candle
		candle.ProductCode = productCode
		candle.Duration = duration
		rows.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return dfCandle, nil
}
