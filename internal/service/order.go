package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/trm"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type OrderRepo interface {
	GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error)
	LatestOrders(ctx context.Context, count int) ([]entities.Order, error)

	// Операции идемпотентны, т.к. используется ON CONFLICT DO NOTHING
	SaveItems(ctx context.Context, orderUID string, items []entities.Item) error
	SavePayment(ctx context.Context, orderUID string, p entities.Payment) error
	SaveDelivery(ctx context.Context, orderUID string, d entities.Delivery) error
	SaveOrder(ctx context.Context, o entities.Order) error
}

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte)
}

type OrderService struct {
	logger    *slog.Logger
	txManager trm.Manager
	repo      OrderRepo
	cache     Cache
}

func NewOrderService(logger *slog.Logger, txManager trm.Manager, repo OrderRepo, cache Cache) *OrderService {
	return &OrderService{
		logger:    logger.With(slog.String("service", "order")),
		txManager: txManager,
		repo:      repo,
		cache:     cache,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order entities.Order) error {
	fn := func() error {
		return s.txManager.Do(ctx, func(ctx context.Context) error {
			// для начала надо сохранить информацию о заказе чтобы был доступен внешний ключ
			if err := s.repo.SaveOrder(ctx, order); err != nil {
				return fmt.Errorf("failed to save order: %w", err)
			}

			eg, ctx := errgroup.WithContext(ctx)

			// порядок не имеет значения, поэтому можно сохранять параллельно
			eg.Go(func() error {
				return s.repo.SaveDelivery(ctx, order.OrderUID, order.Delivery)
			})
			eg.Go(func() error {
				return s.repo.SavePayment(ctx, order.OrderUID, order.Payment)
			})
			eg.Go(func() error {
				return s.repo.SaveItems(ctx, order.OrderUID, order.Items)
			})

			if err := eg.Wait(); err != nil {
				return fmt.Errorf("failed to save order : %w", err)
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

	err := utils.Retry(cfg, fn)
	if err != nil {
		return fmt.Errorf("failed after retry: %w", err)
	}
	return nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error) {
	// Проверяем кэш
	if data, ok := s.cache.Get(orderUID); ok {
		s.logger.DebugContext(ctx, "cache hit", "order_uid", orderUID)
		var order entities.Order
		if err := order.Unmarshal(data); err != nil {
			return entities.Order{}, fmt.Errorf("failed to unmarshal order: %w", err)
		}
		return order, nil
	}

	s.logger.DebugContext(ctx, "cache miss", "order_uid", orderUID)
	var order entities.Order
	fn := func() error {
		var err error
		order, err = s.repo.GetOrderByID(ctx, orderUID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		return nil
	}
	cfg := utils.RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxAttempts:  5,
		Multiplier:   2,
	}
	// Не ретраим если заказ не найден
	if err := utils.Retry(cfg, fn, entities.ErrOrderNotFound); err != nil {
		return entities.Order{}, fmt.Errorf("failed after retry: %w", err)
	}

	data, err := order.Marshal()
	if err != nil {
		return entities.Order{}, fmt.Errorf("failed to marshal order: %w", err)
	}
	s.cache.Set(orderUID, data)
	return order, nil
}

func (s *OrderService) WarmUpCache(ctx context.Context, count int) error {
	orders, err := s.repo.LatestOrders(ctx, count)
	if err != nil {
		return fmt.Errorf("failed to get latest orders: %w", err)
	}

	for _, order := range orders {
		data, err := order.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal order: %w", err)
		}
		s.cache.Set(order.OrderUID, data)
	}

	s.logger.InfoContext(ctx, "cache warmed up", slog.Int("count", len(orders)))
	return nil
}
