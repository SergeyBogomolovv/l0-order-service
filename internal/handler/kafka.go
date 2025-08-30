package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/SergeyBogomolovv/l0-order-service/internal/config"
	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
)

type OrderSaver interface {
	SaveOrder(ctx context.Context, order entities.Order) error
}

type KafkaHandler struct {
	dlq      *kafka.Writer
	reader   *kafka.Reader
	logger   *slog.Logger
	validate *validator.Validate
	saver    OrderSaver
}

func NewKafkaHandler(logger *slog.Logger, cfg config.Kafka, saver OrderSaver) *KafkaHandler {
	return &KafkaHandler{
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

func (h *KafkaHandler) Consume(ctx context.Context) {
	for {
		m, err := h.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				break
			}
			h.logger.ErrorContext(ctx, "failed to fetch message", slog.Any("error", err))
			continue
		}

		start := time.Now()

		// В операции сохранения уже есть retry
		err = h.handleSaveOrder(ctx, m)
		if err != nil {
			ordersFailed.Inc()
			h.logger.ErrorContext(ctx, "failed to handle message", slog.Any("error", err))

			// В библиотеке уже есть retry
			err := h.WriteToDLQ(ctx, m)
			if err != nil {
				h.logger.ErrorContext(ctx, "failed to write message to DLQ", slog.Any("error", err))
				continue
			}
			ordersDLQ.Inc()
		} else {
			ordersProcessed.Inc()
		}

		orderProcessingDuration.Observe(time.Since(start).Seconds())

		if err := h.reader.CommitMessages(ctx, m); err != nil {
			commitErrors.Inc()
			h.logger.ErrorContext(ctx, "failed to commit message", slog.Any("error", err))
		}
	}
}

func (h *KafkaHandler) handleSaveOrder(ctx context.Context, m kafka.Message) error {
	var order Order
	if err := json.Unmarshal(m.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	if err := h.validate.Struct(order); err != nil {
		return fmt.Errorf("invalid order data: %w", err)
	}

	return h.saver.SaveOrder(ctx, OrderJSONToEntity(order))
}

func (h *KafkaHandler) WriteToDLQ(ctx context.Context, m kafka.Message) error {
	m.Topic = fmt.Sprintf("%s-dlq", m.Topic)
	return h.dlq.WriteMessages(ctx, m)
}

func (h *KafkaHandler) Close() error {
	if err := h.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	if err := h.dlq.Close(); err != nil {
		return fmt.Errorf("failed to close dlq: %w", err)
	}
	return nil
}
