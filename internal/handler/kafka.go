package handler

import (
	"OrderService/internal/config"
	"OrderService/internal/entities"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
)

type OrderSaver interface {
	SaveOrder(ctx context.Context, order entities.Order) error
}

type kafkaHandler struct {
	dlq      *kafka.Writer
	reader   *kafka.Reader
	logger   *slog.Logger
	validate *validator.Validate
	saver    OrderSaver
}

func NewKafkaHandler(logger *slog.Logger, cfg config.Kafka, saver OrderSaver) *kafkaHandler {
	return &kafkaHandler{
		logger: logger.With(slog.String("handler", "kafka")),
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupID,
			Topic:   cfg.Topic,
			MaxWait: cfg.ReaderMaxWait,
		}),
		dlq: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: cfg.BatchTimeout,
		},
		validate: validator.New(),
		saver:    saver,
	}
}

func (h *kafkaHandler) Consume(ctx context.Context) {
	for {
		m, err := h.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				break
			} else {
				h.logger.Error("failed to fetch message", slog.Any("error", err))
				continue
			}
		}

		// В операции сохранения уже есть retry
		if err := h.handleSaveOrder(ctx, m); err != nil {
			h.logger.Error("failed to handle message", slog.Any("error", err))

			// В библиотеке уже есть retry
			if err := h.WriteToDLQ(ctx, m); err != nil {
				h.logger.Error("failed to write message to DLQ", slog.Any("error", err))
				continue
			}
		}

		if err := h.reader.CommitMessages(ctx, m); err != nil {
			h.logger.Error("failed to commit message", slog.Any("error", err))
		}
	}
}

func (h *kafkaHandler) handleSaveOrder(ctx context.Context, m kafka.Message) error {
	var order Order
	if err := json.Unmarshal(m.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	if err := h.validate.Struct(order); err != nil {
		return fmt.Errorf("invalid order data: %w", err)
	}

	return h.saver.SaveOrder(ctx, OrderJSONToEntity(order))
}

func (h *kafkaHandler) WriteToDLQ(ctx context.Context, m kafka.Message) error {
	m.Topic = fmt.Sprintf("%s-dlq", m.Topic)
	return h.dlq.WriteMessages(ctx, m)
}

func (h *kafkaHandler) Close() error {
	if err := h.reader.Close(); err != nil {
		return err
	}
	return h.dlq.Close()
}
