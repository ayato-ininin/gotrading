package controllers

import (
	"encoding/json"
	"fmt"
	"gotrading/app/models"
	"gotrading/config"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"text/template"
)

//キャッシュができる記法
var templates = template.Must(template.ParseFiles("app/views/google.html"))

func viewChartHandler(w http.ResponseWriter, r *http.Request){
	limit := 100
	duration := "1m"
	durationTime := config.Config.Durations[duration]
	df, _ := models.GetAllCandle(config.Config.ProductCode,durationTime,limit)


	err := templates.ExecuteTemplate(w, "google.html",df.Candles)
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
	}
}

type JSONError struct{
	Error string `json:"error"`
	Code int `json:"code"` //エラーコード
}

//jsonになにかあったときに、jsonで返すapiエラー自作
func APIError(w http.ResponseWriter, errMessage string , code int){
	w.Header().Set("Content-Type","application/json")//レスポンスヘッダ
	w.WriteHeader(code)//エラーコード
	jsonError, err := json.Marshal(JSONError{Error: errMessage,Code:code})
	if err != nil{
		log.Fatal(err)
	}
	w.Write(jsonError)//jsonをreturn
}
//pathチェック用
var apiValidPath = regexp.MustCompile("^/api/candle/$")

//ハンドラーのラップ(section13で解説されている。)
func apiMakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0 {
			APIError(w, "Not found", http.StatusNotFound)
		}
		fn(w, r)
	}
}

//candleのデータをJSONで返すapi
//http://localhost:8080/api/candle/?product_code=BTC_USD&duration=1s
func apiCandleHandler(w http.ResponseWriter, r *http.Request){
	//ajaxで受け取るURLから取る。
	productCode := r.URL.Query().Get("product_code")
	if productCode == ""{
		APIError(w, "no product param", http.StatusBadRequest)
		return
	}
	//ajaxで受け取るURLから取る。
	strLimit := r.URL.Query().Get("limit")
	limit,err := strconv.Atoi(strLimit)
	//最大で1000、0以上、空白は不可で初期値として1000を入れる。
	if strLimit == "" || err != nil || limit<0 || limit >1000{
		limit = 1000
	}

	//これもajaxで受け取るURLからとる。
	duration := r.URL.Query().Get("duration")
	if duration == ""{
		duration = "1m"
	}
	durationTime := config.Config.Durations[duration]

	//DBから取得
	df, _ := models.GetAllCandle(productCode, durationTime, limit)

	//JSONでレスポンスしてあげる。
	json, err := json.Marshal(df)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}


func StartWebServer() error {
	http.HandleFunc("/api/candle/",apiMakeHandler(apiCandleHandler))
	http.HandleFunc("/chart/", viewChartHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d",config.Config.Port),nil)
}
