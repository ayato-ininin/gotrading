package models

import (
	"fmt"
	"gotrading/bitflyer"
	"time"
)

type Candle struct {
	ProductCode string
	Duration time.Duration
	Time time.Time
	Open float64
	Close float64
	High float64
	Low float64
	Volume float64
}

func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open,close,high,low,volume float64) *Candle {
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

//table名を取得する
//1h,1m,1sで共通の処理。
func (c *Candle) TableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

func (c *Candle) Create() error{
	cmd := fmt.Sprintf("INSERT INTO %s (time,open,close,high,low,volume) VALUES (?,?,?,?,?,?)",c.TableName())
	_, err := DbConnection.Exec(cmd, c.Time.Format(time.RFC3339),c.Open,c.Close,c.High,c.Low,c.Volume)
	if err != nil{
		return err
	}
	return err
}

/*
updateメソッド
1h,1m,1sでかぶってるのがあれば、データを更新する。
timestampがkeyとなる。
*/
func (c *Candle) Save() error{
	cmd := fmt.Sprintf("UPDATE %s SET open = ?,close=?,high = ?,low= ?,volume = ? WHERE time = ?",c.TableName())
	_, err := DbConnection.Exec(cmd,c.Open,c.Close,c.High,c.Low,c.Volume,c.Time.Format(time.RFC3339))
	if err != nil{
		return err
	}
	return err
}

func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *Candle{
	tableName := GetCandleTableName(productCode, duration)
	cmd:= fmt.Sprintf("SELECT time, open,close,high,low,volume FROM time = ?",tableName)
	row := DbConnection.QueryRow(cmd,dateTime.Format(time.RFC3339))//一行のみ返す
	var candle Candle
	err := row.Scan(&candle.Time, &candle.Open,&candle.Close,&candle.High,&candle.Low,&candle.Volume)
	if err != nil{
		return nil
	}
	return NewCandle(productCode,duration,candle.Time,candle.Open,candle.Close,candle.High,candle.Low,candle.Volume)
}

/*
リアルタイム通信でtickerが何度も来るから、その度に発火する。
データベースで新たなrowを作成したらtrue
データを作成せず、更新だけならfalse
*/
func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	//まずDBにデータがあるのかを確認。1mなら、1m15sとかもすべて1mとしてtruncateされてくる
	currentCandle := GetCandle(productCode, duration,ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	//データがない場合
	if currentCandle == nil{
		candle := NewCandle(productCode,duration,ticker.TruncateDateTime(duration),
		price,price,price,price,ticker.Volume)
		candle.Create()
		return true
	}
	//あれば更新のために条件分岐
	if currentCandle.High <= price{
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
