package handler_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
	"github.com/SergeyBogomolovv/l0-order-service/internal/handler"
	mocks "github.com/SergeyBogomolovv/l0-order-service/internal/handler/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHTTPHandler_GetOrderByID(t *testing.T) {
	validOrder := entities.Order{OrderUID: "123"}

	testCases := []struct {
		name         string
		orderUID     string
		mockBehavior func(svc *mocks.MockOrderGetter)
		wantStatus   int
		wantBody     string
	}{
		{
			name:     "success",
			orderUID: "123",
			mockBehavior: func(svc *mocks.MockOrderGetter) {
				svc.EXPECT().
					GetOrderByID(mock.Anything, "123").
					Return(validOrder, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   `"order_uid":"123"`,
		},
		{
			name:     "not found",
			orderUID: "not-exist",
			mockBehavior: func(svc *mocks.MockOrderGetter) {
				svc.EXPECT().
					GetOrderByID(mock.Anything, "not-exist").
					Return(entities.Order{}, entities.ErrOrderNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"order not found"`,
		},
		{
			name:     "internal error",
			orderUID: "123",
			mockBehavior: func(svc *mocks.MockOrderGetter) {
				svc.EXPECT().
					GetOrderByID(mock.Anything, "123").
					Return(entities.Order{}, errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"internal server error"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := mocks.NewMockOrderGetter(t)
			tc.mockBehavior(svc)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			h := handler.NewHTTPHandler(logger, svc)

			r := chi.NewRouter()
			h.Init(r)

			req := httptest.NewRequest(http.MethodGet, "/order/"+tc.orderUID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			assert.Contains(t, string(body), tc.wantBody)

			if tc.wantStatus == http.StatusOK {
				var resp map[string]any
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, "123", resp["order_uid"])
			}
		})
	}
}
