package entities

import "time"

type Payment struct {
	Transaction  string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDT    time.Time
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}
