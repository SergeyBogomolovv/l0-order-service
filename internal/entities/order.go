package entities

import (
	"errors"
	"time"
)

type Order struct {
	OrderUID        string
	TrackNumber     string
	Entry           string
	Locale          string
	InternalSig     string
	CustomerID      string
	DeliveryService string
	ShardKey        string
	SmID            int
	DateCreated     time.Time
	OofShard        string

	// тут без указателей, потому что предполагается что эти данные всегда присутствуют
	Delivery Delivery
	Payment  Payment
	Items    []Item
}

var (
	ErrOrderNotFound = errors.New("order not found")
)
