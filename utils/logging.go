package utils

import (
	"io"
	"log"
	"os"
)

func LoggingSettings(logFile string)  {
	//O_RDWR : ファイルの読み込みと書き込み両方
	//O_CREATE: ファイルがなければ作成
	//O_APPEND: 上書きではなく追記する
	//0666: ファイルのパーミッション、「rw-rw-rw-」なら「0666」、自分-グループ-他人
	//r:4(読み)、w:2(書き)、x:1(実行)
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err != nil {
		log.Fatal("file=logFile err=%s", err.Error())
	}
	//Stdout→standard output(標準出力:コンソールに出る！！これがないと、ログファイルにしかかかれない。multiwriterは標準出力とログファイルの両方に書き込むという設定)
	multiLogFile := io.MultiWriter(os.Stdout,logfile)
	//出力時の情報を追加で付加したい場合.
	//定数はビットフラグで定義されているので、| でまとめて設定できます：
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(multiLogFile)//出力先
}
