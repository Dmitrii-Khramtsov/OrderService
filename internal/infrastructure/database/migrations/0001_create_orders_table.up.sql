CREATE TABLE orders (
    order_uid          TEXT PRIMARY KEY NOT NULL,
    track_number       TEXT UNIQUE NOT NULL,
    entry              TEXT NOT NULL,
    locale             TEXT NOT NULL,
    internal_signature TEXT,
    customer_id        TEXT NOT NULL,
    delivery_service   TEXT NOT NULL,
    shardkey           TEXT,
    sm_id              INTEGER NOT NULL,
    date_created       TIMESTAMP NOT NULL,
    oof_shard          TEXT
);

CREATE TABLE delivery (
    order_uid TEXT PRIMARY KEY NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    phone     TEXT NOT NULL CHECK (phone LIKE '+%'),
    zip       TEXT NOT NULL,
    city      TEXT NOT NULL,
    address   TEXT NOT NULL,
    region    TEXT NOT NULL,
    email     TEXT CHECK (email = '' OR email LIKE '%_@__%.__%')
);

CREATE TABLE payment (
    order_uid      TEXT PRIMARY KEY NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction    TEXT NOT NULL,
    request_id     TEXT,
    currency       TEXT NOT NULL,
    provider       TEXT NOT NULL,
    amount         INTEGER NOT NULL CHECK (amount >= 0),
    payment_dt     BIGINT NOT NULL,
    bank           TEXT,
    delivery_cost  INTEGER NOT NULL CHECK (delivery_cost >= 0),
    goods_total    INTEGER NOT NULL CHECK (goods_total >= 0),
    custom_fee     INTEGER NOT NULL CHECK (custom_fee >= 0)
);

CREATE TABLE items (
    chrt_id      INTEGER NOT NULL,
    order_uid    TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    track_number TEXT NOT NULL,
    price        INTEGER NOT NULL CHECK (price >= 0),
    rid          TEXT NOT NULL,
    name         TEXT NOT NULL,
    sale         INTEGER NOT NULL,
    size         TEXT NOT NULL,
    total_price  INTEGER NOT NULL CHECK (total_price >= 0),
    nm_id        INTEGER NOT NULL,
    brand        TEXT NOT NULL,
    status       INTEGER NOT NULL,
    PRIMARY KEY (chrt_id, order_uid)
);

-- Создание индексов
CREATE INDEX idx_orders_track_number ON orders(track_number);
CREATE INDEX idx_items_order_uid ON items(order_uid);
