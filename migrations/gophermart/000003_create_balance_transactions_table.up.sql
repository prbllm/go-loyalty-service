CREATE TABLE gophermart.balance_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES gophermart.users(id) ON DELETE CASCADE,
    order_number VARCHAR(255) NOT NULL,
    sum BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_balance_transactions_user_id ON gophermart.balance_transactions(user_id);
CREATE INDEX idx_balance_transactions_user_processed ON gophermart.balance_transactions(user_id, processed_at DESC);

