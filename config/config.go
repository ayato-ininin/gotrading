package config

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
)

type ConfigList struct {
	ApiKey    string
	ApiSecret string
	LogFile   string
	ProductCode string
}

var Config ConfigList //グローバル変数

// パッケージを読み込むときに、一回だけ読み込まれる。
// main.goからimportされたとき、設定ファイルを読み込むことができる。
// それをグローバル変数に入れてるから、main.goからグローバル変数として呼び出せる仕組み。
// 別途、config.iniファイルの作成が必要。
func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Failed to read file: $v", err)
		os.Exit(1) //configファイルが読めなかったら出る。
	}

	Config = ConfigList{
		ApiKey:    cfg.Section("bitflyer").Key("api_key").String(),
		ApiSecret: cfg.Section("bitflyer").Key("api_secret").String(),
		LogFile:   cfg.Section("gotrading").Key("log_file").String(),
		ProductCode:   cfg.Section("gotrading").Key("product_code").String(),
	}
}
