package dao

type Price struct {
	Symbol          string `bson:"symbol"`
	Price           string `bson:"price"`
	TranscationTime int64  `bson:"transcationTime"`
}

type PriceFilter struct {
	Symbol string `bson:"symbol,omitempty"`
}
