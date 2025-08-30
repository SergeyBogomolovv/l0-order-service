package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/SergeyBogomolovv/l0-order-service/internal/entities"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type OrderGetter interface {
	GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error)
}

type HTTPHandler struct {
	logger   *slog.Logger
	validate *validator.Validate
	svc      OrderGetter
}

func NewHTTPHandler(logger *slog.Logger, svc OrderGetter) *HTTPHandler {
	return &HTTPHandler{
		logger:   logger.With(slog.String("handler", "http")),
		validate: validator.New(),
		svc:      svc,
	}
}

func (h *HTTPHandler) Init(r chi.Router) {
	r.Get("/order/{order_uid}", h.GetOrderByID)
}

// GetOrderByID возвращает заказ по ID.
// @Summary      Получить заказ по UID
// @Description  Возвращает информацию о заказе по его уникальному идентификатору
// @Tags         orders
// @Param        order_uid   path      string  true  "Уникальный идентификатор заказа"
// @Success      200  {object}  Order
// @Failure      400  {object}  utils.ValidationErrorResponse "Ошибка валидации"
// @Failure      404  {object}  utils.ErrorResponse "Заказ не найден"
// @Failure      500  {object}  utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /order/{order_uid} [get]
func (h *HTTPHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderUID := chi.URLParam(r, "order_uid")

	if err := h.validate.Var(orderUID, "required"); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	order, err := h.svc.GetOrderByID(ctx, orderUID)

	if errors.Is(err, entities.ErrOrderNotFound) {
		utils.WriteError(w, "order not found", http.StatusNotFound)
		return
	}

	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get order", slog.Any("error", err), slog.String("orderUID", orderUID))
		utils.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, OrderEntityToJSON(order), http.StatusOK)
}
