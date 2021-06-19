package binance

import (
	"context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/binance-exchange/go-binance"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var b binance.Binance
var point = "https://www.binance.com"
var logger *logrus.Logger

func SetLogger(l *logrus.Logger) {
	logger = l
}

func New(api, secret string) error {

	hmacSigner := &binance.HmacSigner{Key: []byte(secret)}
	ctx, _ := context.WithCancel(context.Background())

	binanceService := binance.NewAPIService(
		point,
		api,
		hmacSigner,
		nil,
		ctx,
	)

	b = binance.NewBinance(binanceService)
	return b.Ping()
}

type rate struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (t rate) price() float64 {
	float, err := strconv.ParseFloat(t.Price, 64)
	if err != nil {
		logger.Errorf("Error ParseFloat() %v", err)
		return 1
	}
	return float
}

func Converter(from, to string, sum float64) (float64, error) {

	u := url.Values{}
	u.Add("symbol", from+to)
	get, err := http.Get("https://api.binance.com/api/v3/ticker/price?" + u.Encode())
	if get.StatusCode != 200 || err != nil {
		logger.Errorf("Error Get() https://api.binance.com/api/v3/ticker/price %v", err)
		return 0, err
	}

	defer get.Body.Close()

	var v rate
	decoder := json.NewDecoder(get.Body)
	err = decoder.Decode(&v)
	if err != nil {
		logger.Errorf("Error Decode() %v", err)
		return 0, err
	}

	return v.price() * sum, nil

}

func Balance() []struct {
	Balance  float64
	Currency string
	Rate     float64
} {

	balance := []struct {
		Balance  float64
		Currency string
		Rate     float64
	}{
		{
			Currency: "BUSD",
			Balance:  0,
			Rate:     0,
		},
	}

	account, err := b.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		logger.Errorf("Error Account() %v", err)
		return nil
	}

	var wg sync.WaitGroup
	for _, b := range account.Balances {

		if b.Free == 0 {
			continue
		}

		wg.Add(1)
		go func(balance []struct {
			Balance  float64
			Currency string
			Rate     float64
		}, b *binance.Balance) {

			rate, _ := Converter(b.Asset, "BUSD", 1)
			balance[0].Balance += b.Free * rate
			wg.Done()

		}(balance, b)

	}

	wg.Wait()

	// rub, _ := Converter("BUSD", "RUB", 1)
	// balance[0].Rate = rub

	return balance
}
