package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID        string   `json:"order_uid"`
	TrackNumber     string   `json:"track_number"`
	Entry           string   `json:"entry"`
	Delivery        Delivery `json:"delivery"`
	Payment         Payment  `json:"payment"`
	Items           []Item   `json:"items"`
	Locale          string   `json:"locale"`
	InternalSign    string   `json:"internal_signature"`
	CustomerID      string   `json:"customer_id"`
	DeliveryService string   `json:"delivery_service"`
	ShardKey        string   `json:"shardkey"`
	SmID            int      `json:"sm_id"`
	DateCreated     string   `json:"date_created"`
	OofShard        string   `json:"oof_shard"`
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateRandomOrder() Order {
	return Order{
		OrderUID:    randomString(16),
		TrackNumber: "TRACK" + randomString(6),
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "John Doe",
			Phone:   fmt.Sprintf("+%d", rand.Intn(9999999999)),
			Zip:     fmt.Sprintf("%06d", rand.Intn(999999)),
			City:    "City" + randomString(4),
			Address: fmt.Sprintf("Street %d", rand.Intn(100)),
			Region:  "Region" + randomString(3),
			Email:   fmt.Sprintf("user%d@example.com", rand.Intn(1000)),
		},
		Payment: Payment{
			Transaction:  randomString(16),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(5000) + 500,
			PaymentDT:    time.Now().Unix(),
			Bank:         "bank" + randomString(4),
			DeliveryCost: rand.Intn(1000),
			GoodsTotal:   rand.Intn(3000),
			CustomFee:    rand.Intn(10),
		},
		Items: []Item{
			{
				ChrtID:      rand.Intn(9999999),
				TrackNumber: "TRACK" + randomString(6),
				Price:       rand.Intn(1000) + 100,
				RID:         randomString(16),
				Name:        "Item " + randomString(5),
				Sale:        rand.Intn(50),
				Size:        fmt.Sprintf("%d", rand.Intn(50)),
				TotalPrice:  rand.Intn(1000),
				NmID:        rand.Intn(999999),
				Brand:       "Brand" + randomString(3),
				Status:      200 + rand.Intn(10),
			},
		},
		Locale:          "en",
		InternalSign:    "",
		CustomerID:      "customer_" + randomString(5),
		DeliveryService: "service" + randomString(4),
		ShardKey:        fmt.Sprintf("%d", rand.Intn(10)),
		SmID:            rand.Intn(999),
		DateCreated:     time.Now().Format(time.RFC3339),
		OofShard:        fmt.Sprintf("%d", rand.Intn(5)),
	}
}

func main() {
	addr := kafka.TCP("localhost:9092")

	writer := &kafka.Writer{
		Addr:  addr,
		Topic: "orders",
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			order := generateRandomOrder()
			data, _ := json.Marshal(order)
			writer.WriteMessages(context.Background(), kafka.Message{Value: data})
			log.Println("order generated", order.OrderUID)
		case <-ctx.Done():
			return
		}
	}
}
