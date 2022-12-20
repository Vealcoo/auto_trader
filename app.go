package autotrader

import (
	"auto_trader/exchange/binance"
	"context"

	"github.com/Vealcoo/go-pkg/notify"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cnf *viper.Viper

func configInit() {
	cnf = viper.New()
	cnf.AddConfigPath("./config")
	cnf.SetConfigName("private")
	cnf.SetConfigType("yaml")
	cnf.AutomaticEnv()

	err := cnf.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

var database *mongo.Database

func dbConn() {
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cnf.GetString("mongo.applyURI")))
	if err != nil {
		panic(err)
	}

	database = mongoClient.Database("auto_trader")
}

var alert *notify.Notify

func alertInit() {
	alert = notify.New()
	alert.SetTelegramNotify(cnf.GetString("notify.TgToken"))
}

func buildClient(cnf *viper.Viper, database *mongo.Database, alert *notify.Notify) {
	binance.BuildClient(cnf, database, alert)
}

func Run() {
	configInit()
	dbConn()
	alertInit()
	buildClient(cnf, database, alert)

	binance.Run()
}
