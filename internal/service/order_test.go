package service_test

import (
	"OrderService/internal/entities"
	"OrderService/internal/service"
	mocks "OrderService/internal/service/mocks"
	txMocks "OrderService/pkg/trm/mocks"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			},
			wantErr: dbError,
		},
		{
			name:  "Retry works (first attempt fails, second succeeds)",
			order: entities.Order{OrderUID: "123"},
			mockBehavior: func(orderRepo *mocks.MockOrderRepo) {
				// первая попытка - SaveOrder падает
				orderRepo.EXPECT().SaveOrder(mock.Anything, mock.Anything).
					Once().Return(fmt.Errorf("temporary error"))
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
