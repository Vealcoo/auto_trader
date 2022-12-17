package dao

type Price struct {
	Symbol          string `bson:"symbol"`
	Price           string `bson:"price"`
	Exchange        string `bson:"exchange"`
	TranscationTime int64  `bson:"transcationTime"`
}

type PriceFilter struct {
	Symbol   string `bson:"symbol,omitempty"`
	Exchange string `bson:"exchange,omitempty"`
}

type Order struct {
	OrderId  int64  `bson:"orderId"`
	Symbol   string `bson:"symbol"`
	Price    string `bson:"price"`
	Quantity string `bson:"quantity"`
	Exchange string `bson:"exchange"`
	Side     string `bson:"side"`
	Check    bool   `bson:"check"`
}

type OrderFilter struct {
	OrderId  int64  `bson:"orderId,omitempty"`
	Exchange string `bson:"exchange,omitempty"`
	Side     string `bson:"side,omitempty"`
	Check    bool   `bson:"check,omitempty"`
}

type OrderUpdate struct {
	Check bool `bson:"check,omitempty"`
}
