package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Dao struct {
	price *mongo.Collection
}

func NewBinanceDao(c *mongo.Database) *Dao {
	return &Dao{
		price: c.Collection("price"),
	}
}

func (dao *Dao) Recode(ctx context.Context, data []*Price) error {
	err := dao.UpdateMany(ctx, data)
	if err != nil {
		err = dao.CreateMany(ctx, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dao *Dao) CreateMany(ctx context.Context, data []*Price) error {
	opt := []interface{}{}
	for _, v := range data {
		opt = append(opt, bson.M{
			"symbol": v.Symbol,
			"price":  v.Price,
		})
	}
	_, err := dao.price.InsertMany(ctx, opt)
	if err != nil {
		return err
	}

	return nil
}

func (dao *Dao) UpdateMany(ctx context.Context, data []*Price) error {
	for _, v := range data {
		_, err := dao.price.UpdateOne(ctx,
			bson.M{
				"symbol": v.Symbol,
			},
			bson.M{
				"$set": bson.M{
					"price": v.Price,
				},
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (dao *Dao) FindOne(ctx context.Context, f *PriceFilter) (data *Price, err error) {
	filter := bson.M{}

	if f.Symbol != "" {
		filter["symbol"] = f.Symbol
	}

	err = dao.price.FindOne(
		ctx,
		filter,
	).Decode(&data)
	if err != nil {
		return nil, err
	}

	return
}
