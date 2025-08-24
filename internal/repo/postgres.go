package repo

import (
	"OrderService/internal/entities"
	"OrderService/pkg/trm"
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type postgresRepo struct {
	db *sqlx.DB
	qb sq.StatementBuilderType
}

func NewPostgresRepo(db *sqlx.DB) *postgresRepo {
	return &postgresRepo{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *postgresRepo) LatestOrders(ctx context.Context, count int) ([]entities.Order, error) {
	query, args := r.qb.Select(
		"order_uid", "track_number", "entry", "locale",
		"internal_signature", "customer_id", "delivery_service",
		"shardkey", "sm_id", "date_created", "oof_shard").
		From("orders").
		OrderBy("date_created DESC").
		Limit(uint64(count)).
		MustSql()

	var orders []Order
	err := r.selectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return []entities.Order{}, nil
	}

	uids := make([]string, len(orders))
	for i, o := range orders {
		uids[i] = o.OrderUID
	}

	query, args = r.qb.Select(
		"order_uid", "name", "phone", "zip",
		"city", "address", "region", "email",
	).
		From("deliveries").
		Where(sq.Eq{"order_uid": uids}).
		MustSql()

	var deliveries []Delivery
	err = r.selectContext(ctx, &deliveries, query, args...)
	if err != nil {
		return nil, err
	}
	deliveryMap := make(map[string]Delivery, len(deliveries))
	for _, d := range deliveries {
		deliveryMap[d.OrderUID] = d
	}

	query, args = r.qb.Select(
		"order_uid", "transaction", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
	).
		From("payments").
		Where(sq.Eq{"order_uid": uids}).
		MustSql()

	var payments []Payment
	err = r.selectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, err
	}
	paymentMap := make(map[string]Payment, len(payments))
	for _, p := range payments {
		paymentMap[p.OrderUID] = p
	}

	query, args = r.qb.Select(
		"order_uid", "chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status",
	).
		From("items").
		Where(sq.Eq{"order_uid": uids}).
		MustSql()

	var items []Item
	err = r.selectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, err
	}
	itemsMap := make(map[string][]Item, len(uids))
	for _, it := range items {
		itemsMap[it.OrderUID] = append(itemsMap[it.OrderUID], it)
	}

	result := make([]entities.Order, 0, len(orders))
	for _, o := range orders {
		d := deliveryMap[o.OrderUID]
		p := paymentMap[o.OrderUID]
		it := itemsMap[o.OrderUID]

		result = append(result, OrderToEntity(o, d, p, it))
	}

	return result, nil
}

func (r *postgresRepo) GetOrderByID(ctx context.Context, orderUID string) (entities.Order, error) {
	query, args := r.qb.Select(
		"order_uid", "track_number", "entry", "locale",
		"internal_signature", "customer_id", "delivery_service",
		"shardkey", "sm_id", "date_created", "oof_shard").
		From("orders").
		Where(sq.Eq{"order_uid": orderUID}).
		MustSql()

	var order Order
	err := r.getContext(ctx, &order, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.Order{}, entities.ErrOrderNotFound
	}
	if err != nil {
		return entities.Order{}, err
	}

	query, args = r.qb.Select(
		"order_uid", "name", "phone", "zip",
		"city", "address", "region", "email").
		From("deliveries").
		Where(sq.Eq{"order_uid": orderUID}).
		MustSql()

	var delivery Delivery
	err = r.getContext(ctx, &delivery, query, args...)
	if err != nil {
		return entities.Order{}, err
	}

	query, args = r.qb.Select(
		"order_uid", "transaction", "request_id", "currency", "provider", "amount",
		"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		From("payments").
		Where(sq.Eq{"order_uid": orderUID}).
		MustSql()

	var payment Payment
	err = r.getContext(ctx, &payment, query, args...)
	if err != nil {
		return entities.Order{}, err
	}

	query, args = r.qb.Select(
		"order_uid", "chrt_id", "track_number", "price", "rid", "name", "sale",
		"size", "total_price", "nm_id", "brand", "status").
		From("items").
		Where(sq.Eq{"order_uid": orderUID}).
		MustSql()

	var items []Item
	err = r.selectContext(ctx, &items, query, args...)
	if err != nil {
		return entities.Order{}, err
	}

	return OrderToEntity(order, delivery, payment, items), nil
}

func (r *postgresRepo) SaveOrder(ctx context.Context, o entities.Order) error {
	query, args := r.qb.Insert("orders").
		Columns(
			"order_uid", "track_number", "entry", "locale",
			"internal_signature", "customer_id", "delivery_service",
			"shardkey", "sm_id", "date_created", "oof_shard",
		).
		Values(
			o.OrderUID, o.TrackNumber, nullString(o.Entry), nullString(o.Locale),
			nullString(o.InternalSig), o.CustomerID, o.DeliveryService,
			nullString(o.ShardKey), o.SmID, o.DateCreated, nullString(o.OofShard),
		).
		Suffix("ON CONFLICT (order_uid) DO NOTHING").
		MustSql()

	_, err := r.execContext(ctx, query, args...)
	return err
}

func (r *postgresRepo) SaveDelivery(ctx context.Context, orderUID string, d entities.Delivery) error {
	query, args := r.qb.Insert("deliveries").
		Columns("order_uid", "name", "phone", "zip", "city", "address", "region", "email").
		Values(orderUID,
			nullString(d.Name),
			nullString(d.Phone),
			nullString(d.ZIP),
			nullString(d.City),
			nullString(d.Address),
			nullString(d.Region),
			nullString(d.Email),
		).
		Suffix("ON CONFLICT (order_uid) DO NOTHING").
		MustSql()

	_, err := r.execContext(ctx, query, args...)
	return err
}

func (r *postgresRepo) SavePayment(ctx context.Context, orderUID string, p entities.Payment) error {
	query, args := r.qb.Insert("payments").
		Columns("order_uid", "transaction", "request_id", "currency", "provider", "amount",
			"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		Values(
			orderUID, p.Transaction, nullString(p.RequestID), p.Currency, p.Provider, p.Amount,
			p.PaymentDT, nullString(p.Bank), p.DeliveryCost, p.GoodsTotal, nullInt32(p.CustomFee),
		).
		Suffix("ON CONFLICT (order_uid) DO NOTHING").
		MustSql()

	_, err := r.execContext(ctx, query, args...)
	return err
}

func (r *postgresRepo) SaveItems(ctx context.Context, orderUID string, items []entities.Item) error {
	if len(items) == 0 {
		return nil
	}

	q := r.qb.Insert("items").
		Columns("rid", "order_uid", "chrt_id", "track_number", "price", "name",
			"sale", "size", "total_price", "nm_id", "brand", "status").
		Suffix("ON CONFLICT (rid) DO NOTHING")

	for _, it := range items {
		q = q.Values(
			it.RID,
			orderUID,
			it.ChrtID,
			it.TrackNumber,
			it.Price,
			it.Name,
			nullInt32(it.Sale),
			nullString(it.Size),
			it.TotalPrice,
			it.NmID,
			nullString(it.Brand),
			it.Status,
		)
	}

	query, args := q.MustSql()
	_, err := r.execContext(ctx, query, args...)
	return err
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt32(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func (r *postgresRepo) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx := trm.ExtractTx(ctx)
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

func (r *postgresRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := trm.ExtractTx(ctx)
	if tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.db.GetContext(ctx, dest, query, args...)
}

func (r *postgresRepo) selectContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := trm.ExtractTx(ctx)
	if tx != nil {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	return r.db.SelectContext(ctx, dest, query, args...)
}
