BEGIN;

CREATE TABLE IF NOT EXISTS orders (
  order_uid TEXT PRIMARY KEY,
  track_number TEXT NOT NULL UNIQUE,
  entry TEXT,
  locale TEXT,
  internal_signature TEXT,
  customer_id TEXT NOT NULL,
  delivery_service TEXT NOT NULL,
  shardkey TEXT,
  sm_id INTEGER NOT NULL,
  date_created TIMESTAMPTZ NOT NULL,
  oof_shard TEXT
);

CREATE TABLE IF NOT EXISTS deliveries (
  order_uid TEXT PRIMARY KEY REFERENCES orders (order_uid) ON DELETE CASCADE, -- будем считать что доставка может быть только одна
  name TEXT,
  phone TEXT,
  zip TEXT,
  city TEXT,
  address TEXT,
  region TEXT,
  email TEXT
);

CREATE TABLE IF NOT EXISTS payments (
  order_uid TEXT PRIMARY KEY REFERENCES orders (order_uid) ON DELETE CASCADE, -- будем считать что оплата может быть только одна
  transaction TEXT NOT NULL,
  request_id TEXT,
  currency TEXT NOT NULL,
  provider TEXT NOT NULL,
  amount INTEGER NOT NULL CHECK (amount >= 0),
  payment_dt TIMESTAMPTZ NOT NULL,
  bank TEXT,
  delivery_cost INTEGER NOT NULL CHECK (delivery_cost >= 0),
  goods_total INTEGER NOT NULL CHECK (goods_total >= 0),
  custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
  rid TEXT PRIMARY KEY,
  order_uid TEXT NOT NULL REFERENCES orders (order_uid) ON DELETE CASCADE,
  chrt_id BIGINT NOT NULL,
  track_number TEXT NOT NULL,
  price INTEGER NOT NULL CHECK (price >= 0),
  name TEXT NOT NULL,
  sale INTEGER,
  size TEXT,
  total_price INTEGER NOT NULL CHECK (total_price >= 0),
  nm_id INTEGER NOT NULL,
  brand TEXT,
  status INTEGER NOT NULL
);

COMMIT;
