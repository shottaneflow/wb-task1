
CREATE TABLE IF NOT EXISTS orders (
    order_uid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard TEXT
);


CREATE TABLE IF NOT EXISTS delivery (
    delivery_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uid UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);


CREATE TABLE IF NOT EXISTS payment (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uid UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount NUMERIC,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost NUMERIC,
    goods_total NUMERIC,
    custom_fee NUMERIC
);


CREATE TABLE IF NOT EXISTS items (
    item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uid UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number TEXT,
    price NUMERIC,
    rid TEXT,
    name TEXT,
    sale NUMERIC,
    size TEXT,
    total_price NUMERIC,
    nm_id BIGINT,
    brand TEXT,
    status INTEGER
);
