package repo

import (
	"OrderService/internal/entities"
	"OrderService/pkg/trm"
	"context"
	"database/sql"

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

func (r *postgresRepo) SaveOrder(ctx context.Context, o entities.Order) error {
	q := r.qb.Insert("orders").
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
		Suffix("ON CONFLICT (order_uid) DO NOTHING")

	sqlStr, args := q.MustSql()

	_, err := r.execContext(ctx, sqlStr, args...)
	return err
}

func (r *postgresRepo) SaveDelivery(ctx context.Context, orderUID string, d entities.Delivery) error {
	q := r.qb.Insert("deliveries").
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
		Suffix("ON CONFLICT (order_uid) DO NOTHING")

	sqlStr, args := q.MustSql()

	_, err := r.execContext(ctx, sqlStr, args...)
	return err
}

func (r *postgresRepo) SavePayment(ctx context.Context, orderUID string, p entities.Payment) error {
	q := r.qb.Insert("payments").
		Columns("order_uid", "transaction", "request_id", "currency", "provider", "amount",
			"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		Values(
			orderUID, p.Transaction, nullString(p.RequestID), p.Currency, p.Provider, p.Amount,
			p.PaymentDT, nullString(p.Bank), p.DeliveryCost, p.GoodsTotal, nullInt32(p.CustomFee),
		).
		Suffix("ON CONFLICT (order_uid) DO NOTHING")

	sqlStr, args := q.MustSql()

	_, err := r.execContext(ctx, sqlStr, args...)
	return err
}

func (r *postgresRepo) SaveItems(ctx context.Context, orderUID string, items []entities.Item) error {
	if len(items) == 0 {
		return nil
	}

	q := r.qb.Insert("items").
		Columns("rid", "order_uid", "chrt_id", "track_number", "price", "name",
			"sale", "size", "total_price", "nm_id", "brand", "status")

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

	q = q.Suffix("ON CONFLICT (rid) DO NOTHING")

	sqlStr, args := q.MustSql()

	_, err := r.execContext(ctx, sqlStr, args...)
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

// func (r *postgresRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
// 	tx := trm.ExtractTx(ctx)
// 	if tx != nil {
// 		return tx.GetContext(ctx, dest, query, args...)
// 	}
// 	return r.db.GetContext(ctx, dest, query, args...)
// }
