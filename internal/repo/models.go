package repo

import (
	"database/sql"
	"time"
)

type Order struct {
	OrderUID          string         `db:"order_uid"`
	TrackNumber       string         `db:"track_number"`
	Entry             sql.NullString `db:"entry"`
	Locale            sql.NullString `db:"locale"`
	InternalSignature sql.NullString `db:"internal_signature"`
	CustomerID        string         `db:"customer_id"`
	DeliveryService   string         `db:"delivery_service"`
	ShardKey          sql.NullString `db:"shardkey"`
	SmID              int            `db:"sm_id"`
	DateCreated       time.Time      `db:"date_created"`
	OofShard          sql.NullString `db:"oof_shard"`
}

type Delivery struct {
	OrderUID string         `db:"order_uid"`
	Name     sql.NullString `db:"name"`
	Phone    sql.NullString `db:"phone"`
	Zip      sql.NullString `db:"zip"`
	City     sql.NullString `db:"city"`
	Address  sql.NullString `db:"address"`
	Region   sql.NullString `db:"region"`
	Email    sql.NullString `db:"email"`
}

type Payment struct {
	OrderUID     string         `db:"order_uid"`
	Transaction  string         `db:"transaction"`
	RequestID    sql.NullString `db:"request_id"`
	Currency     string         `db:"currency"`
	Provider     string         `db:"provider"`
	Amount       int            `db:"amount"`
	PaymentDT    time.Time      `db:"payment_dt"`
	Bank         sql.NullString `db:"bank"`
	DeliveryCost int            `db:"delivery_cost"`
	GoodsTotal   int            `db:"goods_total"`
	CustomFee    sql.NullInt32  `db:"custom_fee"`
}

type Item struct {
	RID         string         `db:"rid"`
	OrderUID    string         `db:"order_uid"`
	ChrtID      int64          `db:"chrt_id"`
	TrackNumber string         `db:"track_number"`
	Price       int            `db:"price"`
	Name        string         `db:"name"`
	Sale        sql.NullInt32  `db:"sale"`
	Size        sql.NullString `db:"size"`
	TotalPrice  int            `db:"total_price"`
	NmID        int            `db:"nm_id"`
	Brand       sql.NullString `db:"brand"`
	Status      int            `db:"status"`
}
