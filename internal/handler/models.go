package handler

import (
	"OrderService/internal/entities"
	"time"
)

// Order представляет заказ
type Order struct {
	OrderUID        string    `json:"order_uid" validate:"required"`
	TrackNumber     string    `json:"track_number" validate:"required"`
	Entry           string    `json:"entry,omitempty"`
	Delivery        Delivery  `json:"delivery" validate:"required"` // по идее не должно быть так что ее нету
	Payment         Payment   `json:"payment" validate:"required"`  // по идее не должно быть так что его нету
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

func DeliveryEntityToJSON(e entities.Delivery) Delivery {
	return Delivery{
		Name:    e.Name,
		Phone:   e.Phone,
		ZIP:     e.ZIP,
		City:    e.City,
		Address: e.Address,
		Region:  e.Region,
		Email:   e.Email,
	}
}

func DeliveryJSONToEntity(h Delivery) entities.Delivery {
	return entities.Delivery{
		Name:    h.Name,
		Phone:   h.Phone,
		ZIP:     h.ZIP,
		City:    h.City,
		Address: h.Address,
		Region:  h.Region,
		Email:   h.Email,
	}
}

func PaymentEntityToJSON(e entities.Payment) Payment {
	return Payment{
		Transaction:  e.Transaction,
		RequestID:    e.RequestID,
		Currency:     e.Currency,
		Provider:     e.Provider,
		Amount:       e.Amount,
		PaymentDT:    e.PaymentDT.Unix(),
		Bank:         e.Bank,
		DeliveryCost: e.DeliveryCost,
		GoodsTotal:   e.GoodsTotal,
		CustomFee:    e.CustomFee,
	}
}

func PaymentJSONToEntity(h Payment) entities.Payment {
	return entities.Payment{
		Transaction:  h.Transaction,
		RequestID:    h.RequestID,
		Currency:     h.Currency,
		Provider:     h.Provider,
		Amount:       h.Amount,
		PaymentDT:    time.Unix(h.PaymentDT, 0),
		Bank:         h.Bank,
		DeliveryCost: h.DeliveryCost,
		GoodsTotal:   h.GoodsTotal,
		CustomFee:    h.CustomFee,
	}
}

func ItemEntityToJSON(e entities.Item) Item {
	return Item{
		ChrtID:      e.ChrtID,
		TrackNumber: e.TrackNumber,
		Price:       e.Price,
		RID:         e.RID,
		Name:        e.Name,
		Sale:        e.Sale,
		Size:        e.Size,
		TotalPrice:  e.TotalPrice,
		NmID:        e.NmID,
		Brand:       e.Brand,
		Status:      e.Status,
	}
}

func ItemJSONToEntity(h Item) entities.Item {
	return entities.Item{
		ChrtID:      h.ChrtID,
		TrackNumber: h.TrackNumber,
		Price:       h.Price,
		RID:         h.RID,
		Name:        h.Name,
		Sale:        h.Sale,
		Size:        h.Size,
		TotalPrice:  h.TotalPrice,
		NmID:        h.NmID,
		Brand:       h.Brand,
		Status:      h.Status,
	}
}

func OrderEntityToJSON(e entities.Order) Order {
	items := make([]Item, 0, len(e.Items))
	for _, it := range e.Items {
		items = append(items, ItemEntityToJSON(it))
	}

	return Order{
		OrderUID:        e.OrderUID,
		TrackNumber:     e.TrackNumber,
		Entry:           e.Entry,
		Locale:          e.Locale,
		InternalSig:     e.InternalSig,
		CustomerID:      e.CustomerID,
		DeliveryService: e.DeliveryService,
		ShardKey:        e.ShardKey,
		SmID:            e.SmID,
		DateCreated:     e.DateCreated,
		OofShard:        e.OofShard,
		Delivery:        DeliveryEntityToJSON(e.Delivery),
		Payment:         PaymentEntityToJSON(e.Payment),
		Items:           items,
	}
}

func OrderJSONToEntity(h Order) entities.Order {
	items := make([]entities.Item, 0, len(h.Items))
	for _, it := range h.Items {
		items = append(items, ItemJSONToEntity(it))
	}

	return entities.Order{
		OrderUID:        h.OrderUID,
		TrackNumber:     h.TrackNumber,
		Entry:           h.Entry,
		Locale:          h.Locale,
		InternalSig:     h.InternalSig,
		CustomerID:      h.CustomerID,
		DeliveryService: h.DeliveryService,
		ShardKey:        h.ShardKey,
		SmID:            h.SmID,
		DateCreated:     h.DateCreated,
		OofShard:        h.OofShard,
		Delivery:        DeliveryJSONToEntity(h.Delivery),
		Payment:         PaymentJSONToEntity(h.Payment),
		Items:           items,
	}
}
