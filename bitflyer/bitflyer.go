package bitflyer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "https://api.bitflyer.com/v1/"

type APIClient struct{
	key string
	secret string
	httpClient *http.Client
}

//ここも、http.clientとか構造体そのものの値を使うために、値のコピーじゃ無理やからポインタを返す
func New(key, secret string) *APIClient{
	apiClient := &APIClient{key, secret,&http.Client{}}//&つけて、ポインタで返す
	return apiClient
}

/*
apiのヘッダーを作成する関数(値型レシーバー)
これは、apiそれ自体を呼びにいくことはないのがあるから、値になってそう。
http.clientとか呼ぶなら、そのものがいいけどそれもないし、
コピーしてそのままスタック終わりに破棄される方がよさげ。
*/
func (api APIClient) header(method, endpoint string, body []byte) map[string]string{
	timestamp:= strconv.FormatInt(time.Now().Unix(),10)//10進数にする
	log.Println(timestamp)
	message := timestamp + method + endpoint + string(body)

	//Hash Based Message Authentication Code(Hmac)
	//ヘッダーに含まれる事が多い。ユーザ認証用の署名。正しいクライアントか判断。
	//HMACでは、APIを操作する場合、Secretキーをハッシュ値として送信します。ハッシュとして送信されるため、万が一情報が漏洩した場合でもSecretキーを読むことができません。
	//下記三行は、ルーチン。
	mac := hmac.New(sha256.New, []byte(api.secret))//api_secretをバイト配列にして、ハッシュ作成
	mac.Write([]byte(message))//送りたいデータを追加する
	sign := hex.EncodeToString(mac.Sum(nil))//ハッシュにnilを足してからhexにエンコーディングする。
	return map[string]string{
		"ACCESS-KEY" : api.key,
		"ACCESS-TIMESTAMP": timestamp,
		"ACCESS-SIGN": sign,
		"Content-Type": "application/json",
	}
}

//構造体そのものがもつhttp.clientを呼ぶために、ポインタレシーバにしてそう。
func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte) (body []byte, err error){
	baseURL, err := url.Parse(baseURL)//正しいURLかパースは必須(*url.URLの型になる、URL型の構造体を作成)
	if err != nil {
		return
	}
	apiURL, err := url.Parse(urlPath)//正しいURLかパースは必須(*url.URLの型になる、URL型の構造体を作成)
	if err != nil{
		return
	}
	//ResolveReferenceは、*url.URLのメソッド。URLをくっつけてくれる。
	endpoint := baseURL.ResolveReference(apiURL).String()
	log.Printf("action=doRequest endpoint=%s", endpoint)
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))//getの時はdataはnil
	if err != nil{
		return
	}
	//queryがあれば突っ込む
	q := req.URL.Query()//map型
	for key,value := range query{
		q.Add(key,value)
	}
	//エンコードしていれなおさないといけない
	req.URL.RawQuery = q.Encode()//mapのqueryがstringに変わる。

	//mapで返されたヘッダーを入れていく。
	for key, value := range api.header(method, req.URL.RequestURI(),data){
		req.Header.Add(key,value)
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil,err
	}
	//https://pkg.go.dev/net/http#pkg-overview
	//ドキュメントによると、使い終わったらresponseの接続を閉じないといけない。
	//https://qiita.com/stk0724/items/dc400dccd29a4b3d6471
	//閉じないとtcpコネクションが閉じられない可能性。
	defer resp.Body.Close()
	fmt.Println(resp.Body)
	body,err = ioutil.ReadAll(resp.Body)//jsonがバイト配列に変換される
	if err != nil{
		return nil, err
	}
	return body,nil
}

type Balance struct{
	CurrentCode string `json:"currency_code"`
	Amount float64 `json:"amount"`
	Available float64 `json:"available"`
}

/*
*APIClientになってるのは、main.goで呼ばれるときは、ポインタ型だから。
ポインタレシーバ
ポインタ型に対してあるメソッドが定義されているときに、値型変数からそのメソッドを呼び出そうとすると、コンパイラが暗黙的にポインタ型のメソッド呼び出しに変換してくれます。
特にデータ量の大きな構造体に値レシーバのメソッドを定義すると、メソッド呼び出しごとにコピーが発生するので非常に非効率であることがわかります。このことから、構造体におけるメソッド定義は原則ポインタレシーバに対しておこなったほうがよいです。
値レシーバの場合、値そのものがまるっとコピーされるので、メソッド内でいくら値を書き換えても元のレシーバの値にはまったく影響がありません。
レシーバの内部状態を変更したいメソッドは、（参照型を除き）必ずポインタレシーバで定義しなければなりません。
https://skatsuta.github.io/2015/12/29/value-receiver-pointer-receiver/
https://cloudsmith.co.jp/blog/backend/go/2021/06/1816290.html

//おそらく、doRequestの中でhttp.clientをポインタで呼びたいから、
それをポインタで渡すためにここもポインタレシーバになってそう。
*/
func (api *APIClient) GetBalance() ([]Balance, error){
	url:= "me/getbalance"
	resp,err := api.doRequest("GET", url,map[string]string{},nil)
	//log.Printf("url=%s resp=%s", url, string(resp))
	if err != nil{
		log.Printf("acrion=GetBalance err=%s", err.Error())
		return nil, err
	}
	var balance []Balance
	err = json.Unmarshal(resp,&balance)//jsonを構造体に変換してくれる
	if err != nil {
		log.Printf("acrion=GetBalance err=%s", err.Error())
		return nil, err
	}
	return balance, nil
}


/*
https://mholt.github.io/json-to-go/
上記で、jsonをgoの構造体に変えてくれる
apiのレスポンスがドキュメントにあるなら、簡単に対応する構造体が作れる。
*/
type Ticker struct {
	ProductCode     string  `json:"product_code"`
	State           string  `json:"state"`
	Timestamp       string  `json:"timestamp"`
	TickID          int     `json:"tick_id"`
	BestBid         float64     `json:"best_bid"`
	BestAsk         float64     `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64     `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64     `json:"total_ask_depth"`
	MarketBidSize   float64     `json:"market_bid_size"`
	MarketAskSize   float64     `json:"market_ask_size"`
	Ltp             float64     `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

//売りと買いの中間の値を取得
func (t *Ticker) GetMidPrice() float64{
	return (t.BestBidSize + float64(t.BestAsk)) / 2
}

/*
APIから帰ってきたTimestampをデータに入れるとき、
対応しているtimestampに変える必要があるので、そのメソッド

ちなみに、pubnubからの返り値じゃないと、
2022-11-23T01:00:38.947(/tickerから返ってくるtimestamp)
これにゾーン情報がないからParseエラーになる。
*/
func (t *Ticker) DateTime() time.Time{
	dateTime, err := time.Parse(time.RFC3339, t.Timestamp)
	if err != nil {
		log.Printf("action=DateTime, err=%s", err.Error())
	}
	return dateTime
}

/*
dataTimeでParseしたタイムスタンプに対し、
Time型にはTruncateメソッドが用意されていて、指定した大きさ以下の時刻を切り捨てる。
https://pkg.go.dev/time#Duration.Truncate
	trunc := []time.Duration{
		time.Nanosecond,
		time.Microsecond,
		time.Millisecond,
		time.Second,
		2 * time.Second,
		time.Minute,
		10 * time.Minute,
		time.Hour,
	}
hourにしたら、12:10:12→12:00:00になる。
*/
func (t *Ticker) TruncateDateTime(duration time.Duration)  time.Time{
	return t.DateTime().Truncate(duration)
}

/*
getbalanceから汎用的に使う。
doRequestに送るとき、product_codeはbodyではなく、クエリパラメータとして追加
*/
func (api *APIClient) GetTicker(productCode string) (*Ticker, error){
	url:= "ticker"
	resp,err := api.doRequest("GET", url,map[string]string{"product_code":productCode},nil)
	if err != nil{
		return nil, err
	}
	var ticker Ticker
	err = json.Unmarshal(resp,&ticker)//jsonを構造体に変換してくれる
	if err != nil {
		return nil, err
	}
	return &ticker, nil
}
