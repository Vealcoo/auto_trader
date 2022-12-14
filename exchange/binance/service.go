package binance

import (
	"auto_trader/dao"
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Vealcoo/go-pkg/notify"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/adshao/go-binance/v2"
	"github.com/spf13/viper"
)

var apiKey, secretKey string
var client *binance.Client
var checkList []string

var db *dao.Dao
var alert *notify.Notify

func BuildClient(cnf *viper.Viper, database *mongo.Database, n *notify.Notify) {
	apiKey = cnf.GetString("binance.apiKey")
	secretKey = cnf.GetString("binance.secretKey")
	client = binance.NewClient(apiKey, secretKey)
	checkList = cnf.GetStringSlice("checkList")

	db = dao.NewDao(database)
	alert = n
}

func Run() {
	log.Print("Binance auto_trader start...")

	go priceRecorder()
	go anchoredPurchaser()
	go klinePurchaser()
	go orderManger()
	go seller()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sc)

	<-sc
	log.Print("Binance auto_trader close...")
}

func priceRecorder() {
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	fristIn := true
	ctx := context.Background()
	for {
		if fristIn {
			recorder(ctx)
			fristIn = false
		}
		select {
		case <-ticker.C:
			recorder(ctx)
		}
	}
}

func recorder(ctx context.Context) {
	prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
	if err != nil {
		log.Error().Msg(err.Error())
	}

	var data []*dao.Price
	for _, price := range prices {
		data = append(data, &dao.Price{
			Symbol:   price.Symbol,
			Price:    price.Price,
			Exchange: "binance",
		})
	}

	err = db.UpdatePrices(ctx, data)
	if err != nil {
		log.Error().Msg(err.Error())
	}
}

func anchoredPurchaser() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	var usdt float64

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			res, err := client.NewGetAccountService().Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}
			for _, v := range res.Balances {
				if v.Asset == "USDT" {
					if v, err := strconv.ParseFloat(v.Free, 64); err == nil {
						usdt = v
					}
				}
			}

			// Get the current prices of assets on Binance
			prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}

			for _, price := range prices {
				data, err := db.FindPrice(ctx,
					&dao.PriceFilter{
						Symbol:   price.Symbol,
						Exchange: "binance",
					})
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				if time.Now().Unix()-data.TranscationTime < 1800 {
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

				quantity := usdt / nowPrice / 10

				if (nowPrice-anchoredPrice)/anchoredPrice < -0.03 {
					order, err := client.NewCreateOrderService().Symbol(price.Symbol).
						Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(strconv.FormatFloat(quantity, 'f', 2, 64)).
						Price(price.Price).Do(ctx)
					if err != nil {
						log.Error().Msgf("anchoredPurchaser, create order error:", err.Error())
						continue
					}
					log.Info().Msgf("anchoredPurchaser, order:", order)
					alert.NewTgClient().Notify(order, 515597060)

					err = db.UpdatePrice(ctx, &dao.Price{
						Symbol:          data.Symbol,
						Exchange:        data.Exchange,
						TranscationTime: time.Now().Unix(),
					})
					if err != nil {
						log.Error().Msg(err.Error())
						continue
					}
				}
			}
		}
	}
}

func klinePurchaser() {
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()

	var usdt float64

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			res, err := client.NewGetAccountService().Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}
			for _, v := range res.Balances {
				if v.Asset == "USDT" {
					if v, err := strconv.ParseFloat(v.Free, 64); err == nil {
						usdt = v
					}
				}
			}

			// Get the current prices of assets on Binance
			prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}

			for _, price := range prices {
				data, err := db.FindPrice(ctx,
					&dao.PriceFilter{
						Symbol:   price.Symbol,
						Exchange: "binance",
					})
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				if time.Now().Unix()-data.TranscationTime < 1800 {
					continue
				}

				kline, err := client.NewKlinesService().Interval("8h").Limit(1).Symbol(price.Symbol).Do(ctx)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				klineLowPrice, err := strconv.ParseFloat(kline[0].Low, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				klineHighPrice, err := strconv.ParseFloat(kline[0].High, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				avgPrice := (klineLowPrice + klineHighPrice) / 2

				nowPrice, err := strconv.ParseFloat(price.Price, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				quantity := usdt / nowPrice / 10

				if (nowPrice-avgPrice)/avgPrice < -0.05 {
					order, err := client.NewCreateOrderService().Symbol(price.Symbol).
						Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(strconv.FormatFloat(quantity, 'f', 2, 64)).
						Price(price.Price).Do(ctx)
					if err != nil {
						log.Error().Msgf("klinePurchaser, create order error:", err)
						continue
					}
					log.Info().Msgf("klinePurchaser, order:", order)
					alert.NewTgClient().Notify(order, 515597060)

					err = db.UpdatePrice(ctx, &dao.Price{
						Symbol:          data.Symbol,
						Exchange:        data.Exchange,
						TranscationTime: time.Now().Unix(),
					})
					if err != nil {
						log.Error().Msg(err.Error())
						continue
					}
				}
			}
		}
	}
}

func orderManger() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			for _, symbol := range checkList {
				orders, err := client.NewListOrdersService().Symbol(symbol).Do(ctx)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				if len(orders) == 0 {
					continue
				}

				prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				var pricesMap = make(map[string]string)
				for _, price := range prices {
					pricesMap[price.Price] = price.Price
				}

				for _, order := range orders {
					if order.Status == binance.OrderStatusTypeFilled || order.Status == binance.OrderStatusTypePartiallyFilled {
						var side string
						if order.Side == binance.SideTypeBuy {
							side = "buy"
						} else if order.Side == binance.SideTypeSell {
							side = "sell"
						}

						err = db.CreateOrder(ctx,
							&dao.Order{
								OrderId:  order.OrderID,
								Symbol:   order.Symbol,
								Price:    order.Price,
								Quantity: order.ExecutedQuantity,
								Exchange: "binance",
								Side:     side,
							})
						if err != nil {
							log.Error().Msg(err.Error())
							continue
						}
					}
				}
			}
		}
	}
}

func seller() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			orders, err := db.FindOrder(ctx, &dao.OrderFilter{Side: "buy", Exchange: "binance", Check: false})
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}
			if len(orders) == 0 {
				continue
			}

			prices, err := client.NewListPricesService().Symbols(checkList).Do(ctx)
			if err != nil {
				log.Error().Msg(err.Error())
				continue
			}

			var pricesMap = make(map[string]string)
			for _, price := range prices {
				pricesMap[price.Symbol] = price.Price
			}

			for _, order := range orders {
				orderPrice, err := strconv.ParseFloat(order.Price, 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				nowPrice, err := strconv.ParseFloat(pricesMap[order.Symbol], 64)
				if err != nil {
					log.Error().Msg(err.Error())
					continue
				}

				if (nowPrice-orderPrice)/orderPrice > 0.05 {
					sellOrder, err := client.NewCreateOrderService().Symbol(order.Symbol).
						Side(binance.SideTypeSell).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(order.Quantity).
						Price(pricesMap[order.Symbol]).Do(ctx)
					if err != nil {
						log.Error().Msgf("seller, create order error:", err)
						continue
					}
					log.Info().Msgf("seller, order:", sellOrder)
					alert.NewTgClient().Notify(sellOrder, 515597060)

					err = db.UpdateOrder(ctx, order.OrderId, order.Exchange, &dao.OrderUpdate{Check: true})
					if err != nil {
						log.Error().Msg(err.Error())
						continue
					}
				}
			}
		}
	}
}
