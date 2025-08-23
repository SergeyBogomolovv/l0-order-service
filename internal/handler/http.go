package handler

import (
	"OrderService/internal/entities"
	"OrderService/pkg/utils"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type OrderGetter interface {
	GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error)
}

type httpHandler struct {
	logger   *slog.Logger
	validate *validator.Validate
	svc      OrderGetter
}

func NewHttpHandler(logger *slog.Logger, svc OrderGetter) *httpHandler {
	return &httpHandler{
		logger:   logger.With(slog.String("handler", "http")),
		validate: validator.New(),
		svc:      svc,
	}
}

func (h *httpHandler) Init(r chi.Router) {
	r.Get("/order/{order_uid}", h.GetOrderByID)
}

// @Summary      Получить заказ по UID
// @Description  Возвращает информацию о заказе по его уникальному идентификатору
// @Tags         orders
// @Param        order_uid   path      string  true  "Уникальный идентификатор заказа"
// @Success      200  {object}  Order
// @Failure      400  {object}  utils.ValidationErrorResponse "Ошибка валидации"
// @Failure      404  {object}  utils.ErrorResponse "Заказ не найден"
// @Failure      500  {object}  utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /order/{order_uid} [get]
func (h *httpHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "order_uid")

	if err := h.validate.Var(orderUID, "required"); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	order, err := h.svc.GetOrderByID(r.Context(), orderUID)

	if errors.Is(err, entities.ErrOrderNotFound) {
		utils.WriteError(w, "order not found", http.StatusNotFound)
		return
	}

	if err != nil {
		h.logger.Error("failed to get order", slog.Any("error", err), slog.String("orderUID", orderUID))
		utils.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := utils.WriteJSON(w, OrderEntityToJSON(order), http.StatusOK); err != nil {
		h.logger.Error("failed to write response", slog.Any("error", err))
	}
}
