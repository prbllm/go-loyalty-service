CREATE TABLE gophermart.orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES gophermart.users(id) ON DELETE CASCADE,
    number VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'NEW',
    accrual BIGINT DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_orders_user_id ON gophermart.orders(user_id);
CREATE INDEX idx_orders_number ON gophermart.orders(number);
CREATE INDEX idx_orders_status ON gophermart.orders(status);
CREATE INDEX idx_orders_user_uploaded ON gophermart.orders(user_id, uploaded_at DESC);

