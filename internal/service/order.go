package service

import (
	"OrderService/internal/entities"
	"OrderService/pkg/trm"
	"OrderService/pkg/utils"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type OrderRepo interface {
	// Операции идемпотентны, т.к. используется ON CONFLICT DO NOTHING
	SaveItems(ctx context.Context, orderUID string, items []entities.Item) error
	SavePayment(ctx context.Context, orderUID string, p entities.Payment) error
	SaveDelivery(ctx context.Context, orderUID string, d entities.Delivery) error
	SaveOrder(ctx context.Context, o entities.Order) error
}

type orderService struct {
	logger    *slog.Logger
	txManager trm.Manager
	repo      OrderRepo
}

func NewOrderService(logger *slog.Logger, txManager trm.Manager, repo OrderRepo) *orderService {
	return &orderService{
		logger:    logger.With(slog.String("service", "order")),
		txManager: txManager,
		repo:      repo,
	}
}

func (s *orderService) SaveOrder(ctx context.Context, order entities.Order) error {
	fn := func() error {
		return s.txManager.Do(ctx, func(ctx context.Context) error {
			if err := s.repo.SaveOrder(ctx, order); err != nil {
				return fmt.Errorf("failed to save order: %w", err)
			}
			if err := s.repo.SaveDelivery(ctx, order.OrderUID, order.Delivery); err != nil {
				return fmt.Errorf("failed to save delivery: %w", err)
			}
			if err := s.repo.SavePayment(ctx, order.OrderUID, order.Payment); err != nil {
				return fmt.Errorf("failed to save payment: %w", err)
			}
			if err := s.repo.SaveItems(ctx, order.OrderUID, order.Items); err != nil {
				return fmt.Errorf("failed to save items: %w", err)
			}

			s.logger.Debug("order saved", "order_uid", order.OrderUID)
			return nil
		})
	}

	cfg := utils.RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxAttempts:  5,
		Multiplier:   2,
	}

	return utils.Retry(cfg, fn)
}

func (s *orderService) GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error) {
	return entities.Order{}, nil
}
