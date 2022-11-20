package main

import (
	"fmt"
	"gotrading/bitflyer"
	"gotrading/config"
	"gotrading/utils"
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
	fmt.Println(apiClient.GetBalance())//ポインタのメソッド
}
