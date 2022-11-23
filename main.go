package main

import (
	"fmt"
	"gotrading/bitflyer"
	"gotrading/config"
	"gotrading/utils"
	"time"
)

/*
vscodeでimport自動追加とか定義移動の補完付ける場合、setting.jsonの修正がいる。
また、vscodeからgoplsのインストールのおすすめくるけど、おそらくそこから押しても
go getになるから、go1.16以上とかだと使えないぽい。
go install golang.org/x/tools/gopls@latest
*/

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)

	tickerChannel := make(chan bitflyer.Ticker)
	go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)
	/*channelは受け取ると値が消える。配列的ではない。
	GetRealTimeTickerに「bitflyer.Ticker」の型のデータが送られるチャネルを作っておくると、
	裏でそのチャネルにどんどん突っ込まれるから、それを読んでいく流れ。
	チャネルはキューみたい。アンバッファと言われる
	*/
	//しかもここはブロッキング。入ってきたら下の行に進む。ここはforループなので下には行かない。
	//チャネル自体、並行実行と組み合わせで、並行実行された値をブロッキングでawait的に待てる。
	//rangeを使うことでgorutineからのデータを待ち続ける。
	for ticker := range tickerChannel{
		fmt.Println(ticker)
		fmt.Println(ticker.GetMidPrice())
		fmt.Println(ticker.DateTime())
		fmt.Println(ticker.TruncateDateTime(time.Second))
		fmt.Println(ticker.TruncateDateTime(time.Minute))
		fmt.Println(ticker.TruncateDateTime(time.Hour))
	}
}
