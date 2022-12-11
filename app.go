package autotrader

import (
	"auto_trader/exchange/binance"
	"context"

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

func buildClient(cnf *viper.Viper, database *mongo.Database) {
	binance.BuildClient(cnf, database)
}

func Run() {
	configInit()
	dbConn()
	buildClient(cnf, database)

	binance.Run()
}
