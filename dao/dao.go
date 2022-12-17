package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Dao struct {
	price *mongo.Collection
	order *mongo.Collection
}

func NewDao(c *mongo.Database) *Dao {
	return &Dao{
		price: c.Collection("price"),
		order: c.Collection("order"),
	}
}

func (dao *Dao) CreatePrices(ctx context.Context, data []*Price) error {
	opt := []interface{}{}
	for _, v := range data {
		opt = append(opt, bson.M{
			"symbol":   v.Symbol,
			"price":    v.Price,
			"exchange": v.Exchange,
		})
	}
	_, err := dao.price.InsertMany(ctx, opt)
	if err != nil {
		return err
	}

	return nil
}

func (dao *Dao) CreatePrice(ctx context.Context, data *Price) error {
	_, err := dao.price.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (dao *Dao) UpdatePrices(ctx context.Context, data []*Price) error {
	for _, v := range data {
		res, err := dao.price.UpdateOne(ctx,
			bson.M{
				"symbol":   v.Symbol,
				"exchange": v.Exchange,
			},
			bson.M{
				"$set": bson.M{
					"price": v.Price,
				},
			})
		if err != nil || res.MatchedCount == 0 {
			dao.CreatePrice(ctx, v)
		}
	}
	return nil
}

func (dao *Dao) UpdatePrice(ctx context.Context, data *Price) error {
	_, err := dao.price.UpdateOne(ctx,
		bson.M{
			"symbol":   data.Symbol,
			"exchange": data.Exchange,
		},
		bson.M{
			"$set": bson.M{
				"transcationTime": data.TranscationTime,
			},
		})
	if err != nil {
		return err
	}

	return nil
}

func (dao *Dao) FindPrice(ctx context.Context, f *PriceFilter) (data *Price, err error) {
	filter := bson.M{}

	if f.Symbol != "" {
		filter["symbol"] = f.Symbol
	}

	if f.Exchange != "" {
		filter["exchange"] = f.Exchange
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

func (dao *Dao) CreateOrder(ctx context.Context, data *Order) error {
	order, _ := dao.FindOrder(ctx, &OrderFilter{OrderId: data.OrderId, Exchange: data.Exchange})
	if len(order) != 0 {
		return nil
	}

	_, err := dao.order.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (dao *Dao) FindOrder(ctx context.Context, f *OrderFilter) (data []*Order, err error) {
	filter := bson.M{}
	if f.Check != false {
		filter["check"] = f.Check
	}

	if f.OrderId != 0 {
		filter["orderId"] = f.OrderId
	}

	if f.Side != "" {
		filter["side"] = f.Side
	}

	if f.Exchange != "" {
		filter["exchange"] = f.Exchange
	}

	cur, err := dao.order.Find(
		ctx,
		filter,
		options.Find(),
	)
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &data); err != nil {
		return nil, err
	}

	return
}

func (dao *Dao) UpdateOrder(ctx context.Context, orderId int64, exchange string, u *OrderUpdate) error {
	opt := bson.M{}
	if u.Check != false {
		opt["check"] = u.Check
	}

	_, err := dao.order.UpdateOne(
		ctx,
		bson.M{
			"orderId":  orderId,
			"exchange": exchange,
		},
		opt,
	)
	if err != nil {
		return err
	}

	return nil
}
