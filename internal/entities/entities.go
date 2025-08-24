package entities

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

type Delivery struct {
	Name    string
	Phone   string
	ZIP     string
	City    string
	Address string
	Region  string
	Email   string
}

type Item struct {
	ChrtID      int
	TrackNumber string
	Price       int
	RID         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmID        int
	Brand       string
	Status      int
}

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

func (o *Order) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *Order) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(o)
}

func init() {
	gob.Register(Order{})
	gob.Register(Delivery{})
	gob.Register(Payment{})
	gob.Register(Item{})
}
