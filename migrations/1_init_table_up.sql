CREATE TABLE IF NOT EXISTS order(
    order_uid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    track_number TEXT,
    entry TEXT
);
CREATE TABLE IF NOT EXISTS delivery(
    transaction UUID DEFAULT gen_random_uuid(),
    request_id UUID,
    currency

)