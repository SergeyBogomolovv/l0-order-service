package handler

import (
	"time"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
)

// Order представляет заказ
type Order struct {
	OrderUID        string    `json:"order_uid" validate:"required"`
	TrackNumber     string    `json:"track_number" validate:"required"`
	Entry           string    `json:"entry,omitempty"`
	Delivery        Delivery  `json:"delivery" validate:"required"`
	Payment         Payment   `json:"payment" validate:"required"`
	Items           []Item    `json:"items,omitempty" validate:"required,dive"`
	Locale          string    `json:"locale,omitempty"`
	InternalSig     string    `json:"internal_signature,omitempty"`
	CustomerID      string    `json:"customer_id,omitempty"`
	DeliveryService string    `json:"delivery_service,omitempty"`
	ShardKey        string    `json:"shardkey,omitempty"`
	SmID            int       `json:"sm_id,omitempty"`
	DateCreated     time.Time `json:"date_created"`
	OofShard        string    `json:"oof_shard,omitempty"`
}

// Delivery информация о доставке
type Delivery struct {
	Name    string `json:"name,omitempty" validate:"required"`
	Phone   string `json:"phone,omitempty" validate:"required,e164"`
	ZIP     string `json:"zip,omitempty"`
	City    string `json:"city,omitempty"`
	Address string `json:"address,omitempty"`
	Region  string `json:"region,omitempty"`
	Email   string `json:"email,omitempty" validate:"required,email"`
}

// Payment информация об оплате
type Payment struct {
	Transaction  string `json:"transaction,omitempty" validate:"required"`
	RequestID    string `json:"request_id,omitempty"`
	Currency     string `json:"currency,omitempty" validate:"required"`
	Provider     string `json:"provider,omitempty" validate:"required"`
	Amount       int    `json:"amount,omitempty" validate:"gte=0"`
	PaymentDT    int64  `json:"payment_dt,omitempty" validate:"required"`
	Bank         string `json:"bank,omitempty"`
	DeliveryCost int    `json:"delivery_cost,omitempty"`
	GoodsTotal   int    `json:"goods_total,omitempty"`
	CustomFee    int    `json:"custom_fee,omitempty"`
}

// Item товар в заказе
type Item struct {
	ChrtID      int    `json:"chrt_id,omitempty" validate:"required"`
	TrackNumber string `json:"track_number,omitempty" validate:"required"`
	Price       int    `json:"price,omitempty"`
	RID         string `json:"rid,omitempty"`
	Name        string `json:"name,omitempty"`
	Sale        int    `json:"sale,omitempty"`
	Size        string `json:"size,omitempty"`
	TotalPrice  int    `json:"total_price,omitempty"`
	NmID        int    `json:"nm_id,omitempty"`
	Brand       string `json:"brand,omitempty"`
	Status      int    `json:"status,omitempty"`
}

func DeliveryEntityToJSON(d entities.Delivery) Delivery {
	return Delivery{
		Name:    d.Name,
		Phone:   d.Phone,
		ZIP:     d.ZIP,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func DeliveryJSONToEntity(d Delivery) entities.Delivery {
	return entities.Delivery{
		Name:    d.Name,
		Phone:   d.Phone,
		ZIP:     d.ZIP,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func PaymentEntityToJSON(p entities.Payment) Payment {
	return Payment{
		Transaction:  p.Transaction,
		RequestID:    p.RequestID,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       p.Amount,
		PaymentDT:    p.PaymentDT.Unix(),
		Bank:         p.Bank,
		DeliveryCost: p.DeliveryCost,
		GoodsTotal:   p.GoodsTotal,
		CustomFee:    p.CustomFee,
	}
}

func PaymentJSONToEntity(p Payment) entities.Payment {
	return entities.Payment{
		Transaction:  p.Transaction,
		RequestID:    p.RequestID,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       p.Amount,
		PaymentDT:    time.Unix(p.PaymentDT, 0),
		Bank:         p.Bank,
		DeliveryCost: p.DeliveryCost,
		GoodsTotal:   p.GoodsTotal,
		CustomFee:    p.CustomFee,
	}
}

func ItemEntityToJSON(i entities.Item) Item {
	return Item{
		ChrtID:      i.ChrtID,
		TrackNumber: i.TrackNumber,
		Price:       i.Price,
		RID:         i.RID,
		Name:        i.Name,
		Sale:        i.Sale,
		Size:        i.Size,
		TotalPrice:  i.TotalPrice,
		NmID:        i.NmID,
		Brand:       i.Brand,
		Status:      i.Status,
	}
}

func ItemJSONToEntity(i Item) entities.Item {
	return entities.Item{
		ChrtID:      i.ChrtID,
		TrackNumber: i.TrackNumber,
		Price:       i.Price,
		RID:         i.RID,
		Name:        i.Name,
		Sale:        i.Sale,
		Size:        i.Size,
		TotalPrice:  i.TotalPrice,
		NmID:        i.NmID,
		Brand:       i.Brand,
		Status:      i.Status,
	}
}

func OrderEntityToJSON(o entities.Order) Order {
	items := make([]Item, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, ItemEntityToJSON(it))
	}

	return Order{
		OrderUID:        o.OrderUID,
		TrackNumber:     o.TrackNumber,
		Entry:           o.Entry,
		Locale:          o.Locale,
		InternalSig:     o.InternalSig,
		CustomerID:      o.CustomerID,
		DeliveryService: o.DeliveryService,
		ShardKey:        o.ShardKey,
		SmID:            o.SmID,
		DateCreated:     o.DateCreated,
		OofShard:        o.OofShard,
		Delivery:        DeliveryEntityToJSON(o.Delivery),
		Payment:         PaymentEntityToJSON(o.Payment),
		Items:           items,
	}
}

func OrderJSONToEntity(o Order) entities.Order {
	items := make([]entities.Item, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, ItemJSONToEntity(it))
	}

	return entities.Order{
		OrderUID:        o.OrderUID,
		TrackNumber:     o.TrackNumber,
		Entry:           o.Entry,
		Locale:          o.Locale,
		InternalSig:     o.InternalSig,
		CustomerID:      o.CustomerID,
		DeliveryService: o.DeliveryService,
		ShardKey:        o.ShardKey,
		SmID:            o.SmID,
		DateCreated:     o.DateCreated,
		OofShard:        o.OofShard,
		Delivery:        DeliveryJSONToEntity(o.Delivery),
		Payment:         PaymentJSONToEntity(o.Payment),
		Items:           items,
	}
}
