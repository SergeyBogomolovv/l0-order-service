package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
	"github.com/SergeyBogomolovv/l0-order-service/internal/service"
	mocks "github.com/SergeyBogomolovv/l0-order-service/internal/service/mocks"
	txMocks "github.com/SergeyBogomolovv/l0-order-service/pkg/trm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderService_SaveOrder(t *testing.T) {
	type MockBehavior func(orderRepo *mocks.MockOrderRepo)

	dbError := errors.New("db error")

	testCases := []struct {
		name         string
		order        entities.Order
		mockBehavior MockBehavior
		wantErr      error
	}{
		{
			name: "OK",
			order: entities.Order{
				OrderUID: "123",
			},
			mockBehavior: func(orderRepo *mocks.MockOrderRepo) {
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SaveDelivery(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SavePayment(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SaveItems(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "SaveOrder fails",
			order: entities.Order{OrderUID: "123"},
			mockBehavior: func(orderRepo *mocks.MockOrderRepo) {
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).
					Return(dbError)
			},
			wantErr: dbError,
		},
		{
			name:  "SaveDelivery fails",
			order: entities.Order{OrderUID: "123"},
			mockBehavior: func(orderRepo *mocks.MockOrderRepo) {
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SaveDelivery(mock.Anything, mock.Anything, mock.Anything).
					Return(dbError)
				orderRepo.EXPECT().SavePayment(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
				orderRepo.EXPECT().SaveItems(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: dbError,
		},
		{
			name:  "Retry works (first attempt fails, second succeeds)",
			order: entities.Order{OrderUID: "123"},
			mockBehavior: func(orderRepo *mocks.MockOrderRepo) {
				// первая попытка - SaveOrder падает
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).
					Once().Return(errors.New("temporary error"))
				// вторая попытка - всё ок
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).
					Once().Return(nil)
				orderRepo.EXPECT().SaveDelivery(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SavePayment(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderRepo.EXPECT().SaveItems(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orderRepo := mocks.NewMockOrderRepo(t)
			cache := mocks.NewMockCache(t)
			tx := txMocks.NewMockManager(t)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			tx.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(
					func(ctx context.Context, cb func(ctx context.Context) error) error {
						return cb(ctx)
					})

			tc.mockBehavior(orderRepo)

			svc := service.NewOrderService(logger, tx, orderRepo, cache)

			err := svc.SaveOrder(context.Background(), tc.order)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestOrderService_GetOrderByID(t *testing.T) {
	type MockBehavior func(orderRepo *mocks.MockOrderRepo, cache *mocks.MockCache)

	validOrder := entities.Order{OrderUID: "123"}
	validData, err := validOrder.Marshal()
	require.NoError(t, err)

	testCases := []struct {
		name         string
		orderUID     string
		order        entities.Order
		mockBehavior MockBehavior
		wantErr      error
		want         entities.Order
	}{
		{
			name:     "success from cache",
			orderUID: "123",
			order:    validOrder,
			mockBehavior: func(_ *mocks.MockOrderRepo, cache *mocks.MockCache) {
				cache.EXPECT().
					Get("123").
					Return(validData, true).Once()
			},
			want: validOrder,
		},
		{
			name:     "cache hit but unmarshal fails",
			orderUID: "123",
			mockBehavior: func(_ *mocks.MockOrderRepo, cache *mocks.MockCache) {
				cache.EXPECT().
					Get("123").
					Return([]byte("broken"), true).Once()
			},
			wantErr: entities.ErrInvalidOrder,
		},
		{
			name:     "success from repo and set to cache",
			orderUID: "123",
			order:    validOrder,
			mockBehavior: func(orderRepo *mocks.MockOrderRepo, cache *mocks.MockCache) {
				cache.EXPECT().
					Get("123").
					Return(nil, false).Once()
				orderRepo.EXPECT().
					GetOrderByID(mock.Anything, "123").
					Return(validOrder, nil).Once()
				cache.EXPECT().
					Set("123", validData).
					Return().Once()
			},
			want: validOrder,
		},
		{
			name:     "not found in repo",
			orderUID: "not-exist",
			mockBehavior: func(orderRepo *mocks.MockOrderRepo, cache *mocks.MockCache) {
				cache.EXPECT().
					Get("not-exist").
					Return(nil, false).Once()
				orderRepo.EXPECT().
					GetOrderByID(mock.Anything, "not-exist").
					Return(entities.Order{}, entities.ErrOrderNotFound).Once()
			},
			wantErr: entities.ErrOrderNotFound,
		},
		{
			name:     "second attempt from repo",
			orderUID: "123",
			mockBehavior: func(orderRepo *mocks.MockOrderRepo, cache *mocks.MockCache) {
				cache.EXPECT().
					Get("123").
					Return(nil, false).Once()
				orderRepo.EXPECT().
					GetOrderByID(mock.Anything, "123").
					Return(entities.Order{}, errors.New("some error")).Once()
				orderRepo.EXPECT().
					GetOrderByID(mock.Anything, "123").
					Return(validOrder, nil).Once()
				cache.EXPECT().
					Set("123", validData).
					Return().Once()
			},
			want: validOrder,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orderRepo := mocks.NewMockOrderRepo(t)
			cache := mocks.NewMockCache(t)
			tx := txMocks.NewMockManager(t)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			tc.mockBehavior(orderRepo, cache)

			svc := service.NewOrderService(logger, tx, orderRepo, cache)

			got, err := svc.GetOrderByID(context.Background(), tc.orderUID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
