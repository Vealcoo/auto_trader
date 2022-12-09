package binance

import (
	"auto_trader/exchange/binance/dao"
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/delivery"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/spf13/viper"
)

var apiKey, secretKey string

var client *binance.Client
var futuresClient *futures.Client
var deliveryClient *delivery.Client

var checkList []string

var db *dao.Dao

func BuildClient(cnf *viper.Viper, database *mongo.Database) {
	apiKey = cnf.GetString("binance.apiKey")
	secretKey = cnf.GetString("binance.secretKey")

	client = binance.NewClient(apiKey, secretKey)
	futuresClient = binance.NewFuturesClient(apiKey, secretKey)   // USDT-M Futures
	deliveryClient = binance.NewDeliveryClient(apiKey, secretKey) // Coin-M Futures

	checkList = cnf.GetStringSlice("checkList")

	db = dao.NewBinanceDao(database)
}

func Run() {
	log.Print("Binance auto_trader start...")
	priceRecorder()
	anchoredTrader()
}

func anchoredTrader() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	var usdt float64

	ctx := context.Background()
	for {
		priceRecorder()
		select {
		case <-ticker.C:
			res, err := client.NewGetAccountService().Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
			}
			for _, v := range res.Balances {
				if v.Asset == "USDT" {
					log.Print(v)
					if v, err := strconv.ParseFloat(v.Free, 64); err == nil {
						usdt = v
					}
				}
			}
			log.Print("USDT: ", usdt)

			// Get the current prices of assets on Binance
			prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}

			// Check for arbitrage opportunities by comparing the prices of assets on different markets or exchanges
			// (omitted for simplicity)
			for _, price := range prices {
				log.Printf("%s: %s", price.Symbol, price.Price)

				data, err := db.FindOne(ctx, &dao.PriceFilter{Symbol: price.Symbol})
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				anchoredPrice, err := strconv.ParseFloat(data.Price, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}
				nowPrice, err := strconv.ParseFloat(price.Price, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				if time.Now().Unix()-data.TranscationTime < 1800 {
					continue
				}
				if (nowPrice-anchoredPrice)/anchoredPrice < -0.05 {
					// order, err := client.NewCreateOrderService().Symbol(price.Symbol).
					// 	Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
					// 	TimeInForce(binance.TimeInForceTypeGTC).Quantity("5").
					// 	Price("").Do(ctx)
					// if err != nil {
					// 	log.Error().Msg(err.Error())
					// }
					// log.Print("order: ", order)
				}
			}

			// If an arbitrage opportunity is found, execute trades on Binance to take advantage of it
			// (omitted for simplicity)
		}
	}
}

func priceRecorder() {
	ticker := time.NewTicker(1800 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}

			var data []*dao.Price
			for _, price := range prices {
				data = append(data, &dao.Price{
					Symbol: price.Symbol,
					Price:  price.Price,
				})
			}

			err = db.Recode(ctx, data)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}
		}
	}
}
