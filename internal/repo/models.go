package repo

import (
	"database/sql"
	"time"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
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

func DeliveryToEntity(d Delivery) entities.Delivery {
	return entities.Delivery{
		Name:    nullStringToString(d.Name),
		Phone:   nullStringToString(d.Phone),
		ZIP:     nullStringToString(d.Zip),
		City:    nullStringToString(d.City),
		Address: nullStringToString(d.Address),
		Region:  nullStringToString(d.Region),
		Email:   nullStringToString(d.Email),
	}
}

func PaymentToEntity(p Payment) entities.Payment {
	return entities.Payment{
		Transaction:  p.Transaction,
		RequestID:    nullStringToString(p.RequestID),
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       p.Amount,
		PaymentDT:    p.PaymentDT,
		Bank:         nullStringToString(p.Bank),
		DeliveryCost: p.DeliveryCost,
		GoodsTotal:   p.GoodsTotal,
		CustomFee:    nullInt32ToInt(p.CustomFee),
	}
}

func ItemToEntity(i Item) entities.Item {
	return entities.Item{
		ChrtID:      int(i.ChrtID),
		TrackNumber: i.TrackNumber,
		Price:       i.Price,
		RID:         i.RID,
		Name:        i.Name,
		Sale:        nullInt32ToInt(i.Sale),
		Size:        nullStringToString(i.Size),
		TotalPrice:  i.TotalPrice,
		NmID:        i.NmID,
		Brand:       nullStringToString(i.Brand),
		Status:      i.Status,
	}
}

func OrderToEntity(o Order, d Delivery, p Payment, items []Item) entities.Order {
	order := entities.Order{
		OrderUID:        o.OrderUID,
		TrackNumber:     o.TrackNumber,
		Entry:           nullStringToString(o.Entry),
		Locale:          nullStringToString(o.Locale),
		InternalSig:     nullStringToString(o.InternalSignature),
		CustomerID:      o.CustomerID,
		DeliveryService: o.DeliveryService,
		ShardKey:        nullStringToString(o.ShardKey),
		SmID:            o.SmID,
		DateCreated:     o.DateCreated,
		OofShard:        nullStringToString(o.OofShard),
		Delivery:        DeliveryToEntity(d),
		Payment:         PaymentToEntity(p),
	}

	if len(items) > 0 {
		order.Items = make([]entities.Item, 0, len(items))
		for _, it := range items {
			order.Items = append(order.Items, ItemToEntity(it))
		}
	}

	return order
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullInt32ToInt(ni sql.NullInt32) int {
	if ni.Valid {
		return int(ni.Int32)
	}
	return 0
}
